package netcontrol

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"fynescope/control"
	"fynescope/genericps"
	"fynescope/settings"
)

// Client implements control.ScopeController over a TCP connection to a netcontrol.Server.
// All fire-and-forget commands are encoded and sent immediately.
// Synchronous queries block until the matching ResponseMessage arrives (keyed by request ID).
type Client struct {
	conn net.Conn
	enc  *gob.Encoder
	dec  *gob.Decoder

	// nextID is used to generate unique command IDs for synchronous RPCs.
	nextID atomic.Uint64

	// mu guards the encoder (writes) and the pending map.
	mu      sync.Mutex
	pending map[uint64]chan *ResponseMessage

	// Callbacks registered by the GUI layer.
	cbMu               sync.RWMutex
	refreshCallback    func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64)
	refreshEtsCallback func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64)
	bufferCallback     func(size int)
	etsBufferCallback  func(size int)
	displayStatus      func(s string, errorType control.ScopeError)

	closed    chan struct{}
	closeOnce sync.Once
}

// Dial connects to a netcontrol.Server at addr and returns a ready-to-use Client.
// A background goroutine is started to receive telemetry and RPC responses.
func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("netcontrol: Dial %s: %w", addr, err)
	}
	return NewClient(conn), nil
}

// NewClient wraps an already-established net.Conn and starts the receive loop.
func NewClient(conn net.Conn) *Client {
	c := &Client{
		conn:    conn,
		enc:     gob.NewEncoder(conn),
		dec:     gob.NewDecoder(conn),
		pending: make(map[uint64]chan *ResponseMessage),
		closed:  make(chan struct{}),
	}
	// Seed the ID counter so 0 is never used (0 means fire-and-forget).
	c.nextID.Store(1)
	go c.recvLoop()
	return c
}

// Close closes the underlying TCP connection and stops the receive loop.
func (c *Client) Close() error {
	var err error
	c.closeOnce.Do(func() {
		close(c.closed)
		err = c.conn.Close()
	})
	return err
}

// recvLoop reads ServerMessages from the connection and dispatches them.
func (c *Client) recvLoop() {
	for {
		var msg ServerMessage
		if err := c.dec.Decode(&msg); err != nil {
			select {
			case <-c.closed:
				return
			default:
			}
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				slog.Debug("netcontrol client: server closed connection")
			} else {
				slog.Warn("netcontrol client: recv error", "err", err)
			}
			// Cancel all pending RPCs.
			c.mu.Lock()
			for id, ch := range c.pending {
				ch <- &ResponseMessage{ID: id, Success: false, Error: "connection closed"}
				delete(c.pending, id)
			}
			c.mu.Unlock()
			return
		}
		c.dispatch(&msg)
	}
}

func (c *Client) dispatch(msg *ServerMessage) {
	switch msg.Type {
	case TypeResponse:
		rsp, ok := msg.Payload.(*ResponseMessage)
		if !ok {
			slog.Warn("netcontrol client: malformed Response payload", "type", fmt.Sprintf("%T", msg.Payload))
			return
		}
		c.mu.Lock()
		ch, found := c.pending[rsp.ID]
		if found {
			delete(c.pending, rsp.ID)
		}
		c.mu.Unlock()
		if found {
			ch <- rsp
		}

	case TypeRefresh:
		data, ok := msg.Payload.(*DataMessage)
		if !ok {
			return
		}
		c.cbMu.RLock()
		cb := c.refreshCallback
		c.cbMu.RUnlock()
		if cb != nil {
			cb(data.Buffers, data.BuffersMin, data.DigitalBuffers, data.StartTimeOffset, data.XRoundError, data.SamplingTimeInterval)
		}

	case TypeRefreshEts:
		data, ok := msg.Payload.(*EtsDataMessage)
		if !ok {
			return
		}
		c.cbMu.RLock()
		cb := c.refreshEtsCallback
		c.cbMu.RUnlock()
		if cb != nil {
			cb(data.Buffers, data.EtsOutBuffer, data.XRoundError, data.SamplingTimeInterval)
		}

	case TypeBufferResize:
		data, ok := msg.Payload.(*BufferResizeMessage)
		if !ok {
			return
		}
		c.cbMu.RLock()
		cb := c.bufferCallback
		c.cbMu.RUnlock()
		if cb != nil {
			cb(data.Size)
		}

	case TypeEtsBufferResize:
		data, ok := msg.Payload.(*EtsBufferResizeMessage)
		if !ok {
			return
		}
		c.cbMu.RLock()
		cb := c.etsBufferCallback
		c.cbMu.RUnlock()
		if cb != nil {
			cb(data.Size)
		}

	case TypeStatus:
		data, ok := msg.Payload.(*StatusMessage)
		if !ok {
			return
		}
		c.cbMu.RLock()
		cb := c.displayStatus
		c.cbMu.RUnlock()
		if cb != nil {
			cb(data.Text, data.ErrorType)
		}

	default:
		slog.Warn("netcontrol client: unknown server message type", "type", msg.Type)
	}
}

// send encodes a CommandMessage to the server. The caller must not hold c.mu.
func (c *Client) send(cmd *CommandMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := c.enc.Encode(cmd)
	_ = c.conn.SetWriteDeadline(time.Time{})
	return err
}

// rpc sends a command with a unique ID and waits for the matching ResponseMessage.
func (c *Client) rpc(cmdType string, payload any, timeout time.Duration) (*ResponseMessage, error) {
	id := c.nextID.Add(1)
	ch := make(chan *ResponseMessage, 1)

	c.mu.Lock()
	c.pending[id] = ch
	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := c.enc.Encode(&CommandMessage{ID: id, Type: cmdType, Payload: payload})
	_ = c.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("netcontrol rpc %s: encode: %w", cmdType, err)
	}
	c.mu.Unlock()

	select {
	case rsp := <-ch:
		if !rsp.Success {
			return rsp, fmt.Errorf("netcontrol rpc %s: server error: %s", cmdType, rsp.Error)
		}
		return rsp, nil
	case <-time.After(timeout):
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("netcontrol rpc %s: timeout after %v", cmdType, timeout)
	case <-c.closed:
		return nil, fmt.Errorf("netcontrol rpc %s: client closed", cmdType)
	}
}

const defaultRPCTimeout = 10 * time.Second

// --- ScopeController: Fire-and-forget commands ---

func (c *Client) SetTrigger(msg *control.TriggerDescMsg) {
	cmd := &CommandMessage{
		ID:      0,
		Type:    TypeSetTrigger,
		Payload: &SetTriggerCmd{Desc: msg.TriggerDesc},
	}
	if err := c.send(cmd); err != nil {
		slog.Error("netcontrol client: SetTrigger send error", "err", err)
	}
	// Signal done immediately so the caller is not blocked — the server
	// will apply the trigger asynchronously on its side.
	if msg.Done != nil {
		msg.Done <- struct{}{}
	}
}

func (c *Client) SetChannel(msg *settings.ChSettings) {
	ch := *msg
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetChannel, Payload: &SetChannelCmd{Settings: ch}}); err != nil {
		slog.Error("netcontrol client: SetChannel send error", "err", err)
	}
}

func (c *Client) SetDigitalPort(msg *control.DigitalPortMsg) {
	portMsg := *msg
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetDigitalPort, Payload: &SetDigitalPortCmd{Msg: portMsg}}); err != nil {
		slog.Error("netcontrol client: SetDigitalPort send error", "err", err)
	}
}

func (c *Client) SetGenerator(msg *control.GeneratorDescMsg) {
	cmd := &CommandMessage{
		ID:      0,
		Type:    TypeSetGenerator,
		Payload: &SetGeneratorCmd{Desc: msg.GeneratorDesc},
	}
	if err := c.send(cmd); err != nil {
		slog.Error("netcontrol client: SetGenerator send error", "err", err)
	}
	if msg.Done != nil {
		msg.Done <- struct{}{}
	}
}

func (c *Client) SetDemoGen(msg *control.GeneratorDescMsg) {
	cmd := &CommandMessage{
		ID:      0,
		Type:    TypeSetDemoGen,
		Payload: &SetDemoGenCmd{Desc: msg.GeneratorDesc},
	}
	if err := c.send(cmd); err != nil {
		slog.Error("netcontrol client: SetDemoGen send error", "err", err)
	}
	if msg.Done != nil {
		msg.Done <- struct{}{}
	}
}

func (c *Client) SetInterpolationMode(mode settings.InterpolationType) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetInterpolationMode, Payload: &SetInterpolationModeCmd{Mode: mode}}); err != nil {
		slog.Error("netcontrol client: SetInterpolationMode send error", "err", err)
	}
}

func (c *Client) SetScopeScreenWidth(width float64) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetScopeScreenWidth, Payload: &SetScopeScreenWidthCmd{Width: width}}); err != nil {
		slog.Error("netcontrol client: SetScopeScreenWidth send error", "err", err)
	}
}

func (c *Client) SetMaxScreenTime(t float64) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetMaxScreenTime, Payload: &SetMaxScreenTimeCmd{Time: t}}); err != nil {
		slog.Error("netcontrol client: SetMaxScreenTime send error", "err", err)
	}
}

func (c *Client) SuggestSampleCount(sc uint64) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSuggestSampleCount, Payload: &SuggestSampleCountCmd{Count: sc}}); err != nil {
		slog.Error("netcontrol client: SuggestSampleCount send error", "err", err)
	}
}

func (c *Client) SetETSMode() error {
	rsp, err := c.rpc(TypeSetETSMode, &SetETSModeCmd{}, defaultRPCTimeout)
	if err != nil {
		return err
	}
	_ = rsp
	return nil
}

func (c *Client) SetBlockMode() error {
	rsp, err := c.rpc(TypeSetBlockMode, &SetBlockModeCmd{}, defaultRPCTimeout)
	if err != nil {
		return err
	}
	_ = rsp
	return nil
}

func (c *Client) Stop() error {
	rsp, err := c.rpc(TypeStop, &StopCmd{}, defaultRPCTimeout)
	if err != nil {
		return err
	}
	_ = rsp
	return nil
}

func (c *Client) Shutdown() {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeShutdown, Payload: &ShutdownCmd{}}); err != nil {
		slog.Error("netcontrol client: Shutdown send error", "err", err)
	}
}

func (c *Client) SetResolutionMode(mode genericps.RatioMode) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetResolutionMode, Payload: &SetResolutionModeCmd{Mode: mode}}); err != nil {
		slog.Error("netcontrol client: SetResolutionMode send error", "err", err)
	}
}

func (c *Client) SetDigitalPortEnabled(port int, enabled bool) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetDigitalPortEnabled, Payload: &SetDigitalPortEnabledCmd{Port: port, Enabled: enabled}}); err != nil {
		slog.Error("netcontrol client: SetDigitalPortEnabled send error", "err", err)
	}
}

func (c *Client) SetMaxSamplingRate(rate uint32) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetMaxSamplingRate, Payload: &SetMaxSamplingRateCmd{Rate: rate}}); err != nil {
		slog.Error("netcontrol client: SetMaxSamplingRate send error", "err", err)
	}
}

func (c *Client) RequestRestart() {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeRequestRestart, Payload: &RequestRestartCmd{}}); err != nil {
		slog.Error("netcontrol client: RequestRestart send error", "err", err)
	}
}

func (c *Client) SetStreamEnabled(b bool) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetStreamEnabled, Payload: &SetStreamEnabledCmd{Enabled: b}}); err != nil {
		slog.Error("netcontrol client: SetStreamEnabled send error", "err", err)
	}
}

func (c *Client) NewChannels(numberOfChannels int) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeNewChannels, Payload: &NewChannelsCmd{NumberOfChannels: numberOfChannels}}); err != nil {
		slog.Error("netcontrol client: NewChannels send error", "err", err)
	}
}

func (c *Client) SetInfo(info string) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetInfo, Payload: &SetInfoCmd{Info: info}}); err != nil {
		slog.Error("netcontrol client: SetInfo send error", "err", err)
	}
}

func (c *Client) SetScopeModel(model control.ScopeType) {
	if err := c.send(&CommandMessage{ID: 0, Type: TypeSetScopeModel, Payload: &SetScopeModelCmd{Model: model}}); err != nil {
		slog.Error("netcontrol client: SetScopeModel send error", "err", err)
	}
}

func (c *Client) CallBufferCallback(size int) {
	c.cbMu.RLock()
	cb := c.bufferCallback
	c.cbMu.RUnlock()
	if cb != nil {
		cb(size)
	}
}

func (c *Client) CallRefreshCallback(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64) {
	c.cbMu.RLock()
	cb := c.refreshCallback
	c.cbMu.RUnlock()
	if cb != nil {
		cb(buffers, buffersMin, digitalBuffers, startTimeOffset, xRoundError, samplingTimeInterval)
	}
}

func (c *Client) ShowDisplayStatus(s string, errorType control.ScopeError) {
	c.cbMu.RLock()
	cb := c.displayStatus
	c.cbMu.RUnlock()
	if cb != nil {
		cb(s, errorType)
	}
}

func (c *Client) SetDemoDigitalGen(port0Enabled, port1Enabled bool, freq float64, dir genericps.DigitalDemoGenDirection, enc genericps.DigitalDemoGenEncoding, mode genericps.DigitalDemoGenMode, bitDelay float64) error {
	_, err := c.rpc(TypeSetDemoDigitalGen, &SetDemoDigitalGenCmd{
		Port0Enabled: port0Enabled,
		Port1Enabled: port1Enabled,
		Frequency:    freq,
		Direction:    dir,
		Encoding:     enc,
		Mode:         mode,
		BitDelay:     bitDelay,
	}, defaultRPCTimeout)
	return err
}

func (c *Client) SetDemoRlcFilter(channel genericps.ChannelId, genSource genericps.ChannelId, enabled bool, filterType string, r float64, runit string, l float64, lunit string, cval float64, cunit string) error {
	_, err := c.rpc(TypeSetDemoRlcFilter, &SetDemoRlcFilterCmd{
		Channel:    channel,
		GenSource:  genSource,
		Enabled:    enabled,
		FilterType: filterType,
		R:          r,
		RUnit:      runit,
		L:          l,
		LUnit:      lunit,
		C:          cval,
		CUnit:      cunit,
	}, defaultRPCTimeout)
	return err
}

// --- ScopeController: Synchronous query methods ---

func (c *Client) GetInfo() string {
	rsp, err := c.rpc(TypeGetInfo, &GetInfoReq{}, defaultRPCTimeout)
	if err != nil {
		slog.Error("netcontrol client: GetInfo error", "err", err)
		return ""
	}
	if r, ok := rsp.Result.(*GetInfoRsp); ok {
		return r.Info
	}
	return ""
}

func (c *Client) GetScopeModel() control.ScopeType {
	rsp, err := c.rpc(TypeGetScopeModel, &GetScopeModelReq{}, defaultRPCTimeout)
	if err != nil {
		slog.Error("netcontrol client: GetScopeModel error", "err", err)
		return 0
	}
	if r, ok := rsp.Result.(*GetScopeModelRsp); ok {
		return r.Model
	}
	return 0
}

func (c *Client) GetMaxSamplingRate() uint32 {
	rsp, err := c.rpc(TypeGetMaxSamplingRate, &GetMaxSamplingRateReq{}, defaultRPCTimeout)
	if err != nil {
		slog.Error("netcontrol client: GetMaxSamplingRate error", "err", err)
		return 0
	}
	if r, ok := rsp.Result.(*GetMaxSamplingRateRsp); ok {
		return r.Rate
	}
	return 0
}

func (c *Client) GetStreamEnabled() bool {
	rsp, err := c.rpc(TypeGetStreamEnabled, &GetStreamEnabledReq{}, defaultRPCTimeout)
	if err != nil {
		slog.Error("netcontrol client: GetStreamEnabled error", "err", err)
		return false
	}
	if r, ok := rsp.Result.(*GetStreamEnabledRsp); ok {
		return r.Enabled
	}
	return false
}

func (c *Client) GetSamplingTimeInterval() float64 {
	rsp, err := c.rpc(TypeGetSamplingTimeInterval, &GetSamplingTimeIntervalReq{}, defaultRPCTimeout)
	if err != nil {
		slog.Error("netcontrol client: GetSamplingTimeInterval error", "err", err)
		return 0
	}
	if r, ok := rsp.Result.(*GetSamplingTimeIntervalRsp); ok {
		return r.Interval
	}
	return 0
}

func (c *Client) GetTimeBase() uint64 {
	rsp, err := c.rpc(TypeGetTimeBase, &GetTimeBaseReq{}, defaultRPCTimeout)
	if err != nil {
		slog.Error("netcontrol client: GetTimeBase error", "err", err)
		return 0
	}
	if r, ok := rsp.Result.(*GetTimeBaseRsp); ok {
		return r.TimeBase
	}
	return 0
}

func (c *Client) ChannelRanges(chIndex genericps.ChannelId) ([]int32, error) {
	rsp, err := c.rpc(TypeChannelRanges, &ChannelRangesReq{Channel: chIndex}, defaultRPCTimeout)
	if err != nil {
		return nil, err
	}
	r, ok := rsp.Result.(*ChannelRangesRsp)
	if !ok {
		return nil, fmt.Errorf("netcontrol: unexpected ChannelRanges response type %T", rsp.Result)
	}
	return r.Ranges, nil
}

func (c *Client) NumberOfEnabledAnalogChannels() int {
	rsp, err := c.rpc(TypeNumberOfEnabledAnalogChannels, &NumberOfEnabledAnalogChannelsReq{}, defaultRPCTimeout)
	if err != nil {
		slog.Error("netcontrol client: NumberOfEnabledAnalogChannels error", "err", err)
		return 0
	}
	if r, ok := rsp.Result.(*NumberOfEnabledAnalogChannelsRsp); ok {
		return r.Count
	}
	return 0
}

func (c *Client) MinMaxValues() (min, max int32, err error) {
	rsp, rpcErr := c.rpc(TypeMinMaxValues, &MinMaxValuesReq{}, defaultRPCTimeout)
	if rpcErr != nil {
		return 0, 0, rpcErr
	}
	r, ok := rsp.Result.(*MinMaxValuesRsp)
	if !ok {
		return 0, 0, fmt.Errorf("netcontrol: unexpected MinMaxValues response type %T", rsp.Result)
	}
	return r.Min, r.Max, nil
}

func (c *Client) UnitVariantInfo() (info string, err error) {
	rsp, rpcErr := c.rpc(TypeUnitVariantInfo, &UnitVariantInfoReq{}, defaultRPCTimeout)
	if rpcErr != nil {
		return "", rpcErr
	}
	r, ok := rsp.Result.(*UnitVariantInfoRsp)
	if !ok {
		return "", fmt.Errorf("netcontrol: unexpected UnitVariantInfo response type %T", rsp.Result)
	}
	return r.Info, nil
}

func (c *Client) UnitBatchAndSerialInfo() (info string, err error) {
	rsp, rpcErr := c.rpc(TypeUnitBatchAndSerialInfo, &UnitBatchAndSerialInfoReq{}, defaultRPCTimeout)
	if rpcErr != nil {
		return "", rpcErr
	}
	r, ok := rsp.Result.(*UnitBatchAndSerialInfoRsp)
	if !ok {
		return "", fmt.Errorf("netcontrol: unexpected UnitBatchAndSerialInfo response type %T", rsp.Result)
	}
	return r.Info, nil
}

func (c *Client) GetAnalogueOffset(voltageRange int, coupling genericps.Coupling) (maximumVoltage, minimumVoltage float32, err error) {
	rsp, rpcErr := c.rpc(TypeGetAnalogueOffset, &GetAnalogueOffsetReq{VoltageRange: voltageRange, Coupling: coupling}, defaultRPCTimeout)
	if rpcErr != nil {
		return 0, 0, rpcErr
	}
	r, ok := rsp.Result.(*GetAnalogueOffsetRsp)
	if !ok {
		return 0, 0, fmt.Errorf("netcontrol: unexpected GetAnalogueOffset response type %T", rsp.Result)
	}
	return r.Max, r.Min, nil
}

func (c *Client) GetEtsLimits() (maxInterleave, maxCycles int16) {
	rsp, err := c.rpc(TypeGetEtsLimits, &GetEtsLimitsReq{}, defaultRPCTimeout)
	if err != nil {
		slog.Error("netcontrol client: GetEtsLimits error", "err", err)
		return 0, 0
	}
	if r, ok := rsp.Result.(*GetEtsLimitsRsp); ok {
		return r.MaxInterleave, r.MaxCycles
	}
	return 0, 0
}

func (c *Client) IsDemo() bool {
	rsp, err := c.rpc(TypeIsDemo, &IsDemoReq{}, defaultRPCTimeout)
	if err != nil {
		slog.Error("netcontrol client: IsDemo error", "err", err)
		return false
	}
	if r, ok := rsp.Result.(*IsDemoRsp); ok {
		return r.IsDemo
	}
	return false
}

// --- ScopeController: Callback registration ---

func (c *Client) SetRefreshCallback(cb func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64)) {
	c.cbMu.Lock()
	c.refreshCallback = cb
	c.cbMu.Unlock()
}

func (c *Client) SetRefreshEtsCallback(cb func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64)) {
	c.cbMu.Lock()
	c.refreshEtsCallback = cb
	c.cbMu.Unlock()
}

func (c *Client) SetBufferCallback(cb func(size int)) {
	c.cbMu.Lock()
	c.bufferCallback = cb
	c.cbMu.Unlock()
}

func (c *Client) SetEtsBufferCallback(cb func(size int)) {
	c.cbMu.Lock()
	c.etsBufferCallback = cb
	c.cbMu.Unlock()
}

func (c *Client) SetDisplayStatus(cb func(s string, errorType control.ScopeError)) {
	c.cbMu.Lock()
	c.displayStatus = cb
	c.cbMu.Unlock()
}

// --- ScopeController: Connection bypass ---

// GetCon returns nil; the remote client has no direct hardware connection.
func (c *Client) GetCon() *genericps.Connection {
	return nil
}
