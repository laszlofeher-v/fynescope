package netcontrol

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"fynescope/control"
)

// Server coordinates oscilloscope hardware over TCP using the netcontrol protocol.
// It exposes a local ScopeController to remote GUI clients via gob serialization.
type Server struct {
	ctrl     control.ScopeController
	listener net.Listener
	addr     string

	mu        sync.Mutex
	clients   map[*clientSession]struct{}
	closed    chan struct{}
	closeOnce sync.Once
}

type clientSession struct {
	server   *Server
	conn     net.Conn
	enc      *gob.Encoder
	dec      *gob.Decoder
	closedCh chan struct{}

	mu     sync.Mutex
	closed bool
}

func (cs *clientSession) send(msg *ServerMessage) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.closed {
		return net.ErrClosed
	}

	_ = cs.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := cs.enc.Encode(msg)
	_ = cs.conn.SetWriteDeadline(time.Time{})
	return err
}

func (cs *clientSession) close() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if !cs.closed {
		cs.closed = true
		close(cs.closedCh)
		_ = cs.conn.Close()
	}
}

// NewServer creates a new Server instance backed by the given ScopeController.
// It automatically registers callbacks on the controller to broadcast telemetry to connected clients.
func NewServer(ctrl control.ScopeController) *Server {
	if ctrl == nil {
		panic("netcontrol: NewServer called with nil ScopeController")
	}
	s := &Server{
		ctrl:    ctrl,
		clients: make(map[*clientSession]struct{}),
		closed:  make(chan struct{}),
	}
	s.RegisterCallbacks()
	return s
}

// RegisterCallbacks sets up the scope callbacks on the underlying ScopeController
// to stream waveform and status telemetry to all active client connections.
func (s *Server) RegisterCallbacks() {
	s.ctrl.SetRefreshCallback(func(buffers [][]int16, buffersMin [][]int16, digitalBuffers [][]int16, startTimeOffset int64, xRoundError float64, samplingTimeInterval float64) {
		s.broadcast(&ServerMessage{
			Type: TypeRefresh,
			Payload: &DataMessage{
				Buffers:              buffers,
				BuffersMin:           buffersMin,
				DigitalBuffers:       digitalBuffers,
				StartTimeOffset:      startTimeOffset,
				XRoundError:          xRoundError,
				SamplingTimeInterval: samplingTimeInterval,
			},
		})
	})

	s.ctrl.SetRefreshEtsCallback(func(buffers [][]int16, etsOutBuffer []int64, xRoundError float64, samplingTimeInterval float64) {
		s.broadcast(&ServerMessage{
			Type: TypeRefreshEts,
			Payload: &EtsDataMessage{
				Buffers:              buffers,
				EtsOutBuffer:         etsOutBuffer,
				XRoundError:          xRoundError,
				SamplingTimeInterval: samplingTimeInterval,
			},
		})
	})

	s.ctrl.SetBufferCallback(func(size int) {
		s.broadcast(&ServerMessage{
			Type: TypeBufferResize,
			Payload: &BufferResizeMessage{
				Size: size,
			},
		})
	})

	s.ctrl.SetEtsBufferCallback(func(size int) {
		s.broadcast(&ServerMessage{
			Type: TypeEtsBufferResize,
			Payload: &EtsBufferResizeMessage{
				Size: size,
			},
		})
	})

	s.ctrl.SetDisplayStatus(func(text string, errorType control.ScopeError) {
		s.broadcast(&ServerMessage{
			Type: TypeStatus,
			Payload: &StatusMessage{
				Text:      text,
				ErrorType: errorType,
			},
		})
	})
}

// Start listens on the given TCP address and accepts incoming client connections in a background goroutine.
func (s *Server) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.mu.Lock()
	s.listener = ln
	s.addr = ln.Addr().String()
	s.mu.Unlock()

	go func() {
		_ = s.acceptLoop(ln)
	}()

	return nil
}

// Serve accepts client connections on the provided listener until the server is closed.
func (s *Server) Serve(ln net.Listener) error {
	s.mu.Lock()
	s.listener = ln
	s.addr = ln.Addr().String()
	s.mu.Unlock()

	return s.acceptLoop(ln)
}

// Addr returns the listener network address, or nil if not listening.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}

// ClientCount returns the number of currently connected clients.
func (s *Server) ClientCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}

// Close closes the server listener and terminates all active client sessions.
func (s *Server) Close() error {
	var err error
	s.closeOnce.Do(func() {
		close(s.closed)
		s.mu.Lock()
		if s.listener != nil {
			err = s.listener.Close()
		}
		for cs := range s.clients {
			cs.close()
		}
		s.clients = make(map[*clientSession]struct{})
		s.mu.Unlock()
	})
	return err
}

func (s *Server) addClient(cs *clientSession) {
	s.mu.Lock()
	s.clients[cs] = struct{}{}
	s.mu.Unlock()
}

func (s *Server) removeClient(cs *clientSession) {
	cs.close()
	s.mu.Lock()
	delete(s.clients, cs)
	s.mu.Unlock()
}

func (s *Server) broadcast(msg *ServerMessage) {
	s.mu.Lock()
	if len(s.clients) == 0 {
		s.mu.Unlock()
		return
	}
	sessions := make([]*clientSession, 0, len(s.clients))
	for cs := range s.clients {
		sessions = append(sessions, cs)
	}
	s.mu.Unlock()

	for _, cs := range sessions {
		if err := cs.send(msg); err != nil {
			slog.Debug("client session send error during broadcast", "err", err)
			s.removeClient(cs)
		}
	}
}

func (s *Server) acceptLoop(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.closed:
				return nil
			default:
				return err
			}
		}
		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	session := &clientSession{
		server:   s,
		conn:     conn,
		enc:      gob.NewEncoder(conn),
		dec:      gob.NewDecoder(conn),
		closedCh: make(chan struct{}),
	}

	s.addClient(session)
	defer s.removeClient(session)

	for {
		select {
		case <-s.closed:
			return
		case <-session.closedCh:
			return
		default:
		}

		var cmd CommandMessage
		err := session.dec.Decode(&cmd)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				slog.Debug("client disconnected", "remote", conn.RemoteAddr())
			} else {
				select {
				case <-s.closed:
					return
				case <-session.closedCh:
					return
				default:
					slog.Warn("client decode error", "err", err, "remote", conn.RemoteAddr())
				}
			}
			return
		}

		s.handleCommand(session, &cmd)
	}
}

func unwrapPayload[P ~*T, T any](raw any) (T, bool) {
	if raw == nil {
		var zero T
		return zero, false
	}
	if ptr, ok := raw.(P); ok && ptr != nil {
		return *ptr, true
	}
	if val, ok := raw.(T); ok {
		return val, true
	}
	var zero T
	return zero, false
}

func (s *Server) handleCommand(session *clientSession, cmd *CommandMessage) {
	var (
		err    error
		result any
	)

	switch cmd.Type {
	case TypeSetTrigger:
		if p, ok := unwrapPayload[*SetTriggerCmd, SetTriggerCmd](cmd.Payload); ok {
			done := make(chan struct{}, 1)
			msg := control.TriggerDescMsg{
				TriggerDesc: p.Desc,
				Done:        done,
			}
			s.ctrl.SetTrigger(&msg)
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				slog.Warn("SetTrigger timeout waiting for Done channel")
			case <-s.closed:
				return
			case <-session.closedCh:
				return
			}
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetChannel:
		if p, ok := unwrapPayload[*SetChannelCmd, SetChannelCmd](cmd.Payload); ok {
			chCopy := p.Settings
			s.ctrl.SetChannel(&chCopy)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetDigitalPort:
		if p, ok := unwrapPayload[*SetDigitalPortCmd, SetDigitalPortCmd](cmd.Payload); ok {
			portMsg := p.Msg
			s.ctrl.SetDigitalPort(&portMsg)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetGenerator:
		if p, ok := unwrapPayload[*SetGeneratorCmd, SetGeneratorCmd](cmd.Payload); ok {
			done := make(chan struct{}, 1)
			msg := control.GeneratorDescMsg{
				GeneratorDesc: p.Desc,
				Done:          done,
			}
			s.ctrl.SetGenerator(&msg)
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				slog.Warn("SetGenerator timeout waiting for Done channel")
			case <-s.closed:
				return
			case <-session.closedCh:
				return
			}
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetDemoGen:
		if p, ok := unwrapPayload[*SetDemoGenCmd, SetDemoGenCmd](cmd.Payload); ok {
			done := make(chan struct{}, 1)
			msg := control.GeneratorDescMsg{
				GeneratorDesc: p.Desc,
				Done:          done,
			}
			s.ctrl.SetDemoGen(&msg)
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				slog.Warn("SetDemoGen timeout waiting for Done channel")
			case <-s.closed:
				return
			case <-session.closedCh:
				return
			}
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetInterpolationMode:
		if p, ok := unwrapPayload[*SetInterpolationModeCmd, SetInterpolationModeCmd](cmd.Payload); ok {
			s.ctrl.SetInterpolationMode(p.Mode)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetScopeScreenWidth:
		if p, ok := unwrapPayload[*SetScopeScreenWidthCmd, SetScopeScreenWidthCmd](cmd.Payload); ok {
			s.ctrl.SetScopeScreenWidth(p.Width)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetMaxScreenTime:
		if p, ok := unwrapPayload[*SetMaxScreenTimeCmd, SetMaxScreenTimeCmd](cmd.Payload); ok {
			s.ctrl.SetMaxScreenTime(p.Time)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSuggestSampleCount:
		if p, ok := unwrapPayload[*SuggestSampleCountCmd, SuggestSampleCountCmd](cmd.Payload); ok {
			s.ctrl.SuggestSampleCount(p.Count)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetETSMode:
		err = s.ctrl.SetETSMode()

	case TypeSetBlockMode:
		err = s.ctrl.SetBlockMode()

	case TypeStop:
		err = s.ctrl.Stop()

	case TypeShutdown:
		s.ctrl.Shutdown()

	case TypeSetResolutionMode:
		if p, ok := unwrapPayload[*SetResolutionModeCmd, SetResolutionModeCmd](cmd.Payload); ok {
			s.ctrl.SetResolutionMode(p.Mode)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetDigitalPortEnabled:
		if p, ok := unwrapPayload[*SetDigitalPortEnabledCmd, SetDigitalPortEnabledCmd](cmd.Payload); ok {
			s.ctrl.SetDigitalPortEnabled(p.Port, p.Enabled)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetMaxSamplingRate:
		if p, ok := unwrapPayload[*SetMaxSamplingRateCmd, SetMaxSamplingRateCmd](cmd.Payload); ok {
			s.ctrl.SetMaxSamplingRate(p.Rate)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeRequestRestart:
		s.ctrl.RequestRestart()

	case TypeSetStreamEnabled:
		if p, ok := unwrapPayload[*SetStreamEnabledCmd, SetStreamEnabledCmd](cmd.Payload); ok {
			s.ctrl.SetStreamEnabled(p.Enabled)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeNewChannels:
		if p, ok := unwrapPayload[*NewChannelsCmd, NewChannelsCmd](cmd.Payload); ok {
			s.ctrl.NewChannels(p.NumberOfChannels)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetInfo:
		if p, ok := unwrapPayload[*SetInfoCmd, SetInfoCmd](cmd.Payload); ok {
			s.ctrl.SetInfo(p.Info)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetScopeModel:
		if p, ok := unwrapPayload[*SetScopeModelCmd, SetScopeModelCmd](cmd.Payload); ok {
			s.ctrl.SetScopeModel(p.Model)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetDemoDigitalGen:
		if p, ok := unwrapPayload[*SetDemoDigitalGenCmd, SetDemoDigitalGenCmd](cmd.Payload); ok {
			err = s.ctrl.SetDemoDigitalGen(p.Port0Enabled, p.Port1Enabled, p.Frequency, p.Direction, p.Encoding, p.Mode, p.BitDelay)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeSetDemoRlcFilter:
		if p, ok := unwrapPayload[*SetDemoRlcFilterCmd, SetDemoRlcFilterCmd](cmd.Payload); ok {
			err = s.ctrl.SetDemoRlcFilter(p.Channel, p.GenSource, p.Enabled, p.FilterType, p.R, p.RUnit, p.L, p.LUnit, p.C, p.CUnit)
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	// Queries
	case TypeGetAnalogueOffset:
		if p, ok := unwrapPayload[*GetAnalogueOffsetReq, GetAnalogueOffsetReq](cmd.Payload); ok {
			maxV, minV, offsetErr := s.ctrl.GetAnalogueOffset(p.VoltageRange, p.Coupling)
			if offsetErr != nil {
				err = offsetErr
			} else {
				result = &GetAnalogueOffsetRsp{Max: maxV, Min: minV}
			}
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeChannelRanges:
		if p, ok := unwrapPayload[*ChannelRangesReq, ChannelRangesReq](cmd.Payload); ok {
			ranges, rangeErr := s.ctrl.ChannelRanges(p.Channel)
			if rangeErr != nil {
				err = rangeErr
			} else {
				result = &ChannelRangesRsp{Ranges: ranges}
			}
		} else {
			err = fmt.Errorf("invalid payload for %s: %T", cmd.Type, cmd.Payload)
		}

	case TypeMinMaxValues:
		minV, maxV, mmErr := s.ctrl.MinMaxValues()
		if mmErr != nil {
			err = mmErr
		} else {
			result = &MinMaxValuesRsp{Min: minV, Max: maxV}
		}

	case TypeUnitVariantInfo:
		info, uErr := s.ctrl.UnitVariantInfo()
		if uErr != nil {
			err = uErr
		} else {
			result = &UnitVariantInfoRsp{Info: info}
		}

	case TypeUnitBatchAndSerialInfo:
		info, uErr := s.ctrl.UnitBatchAndSerialInfo()
		if uErr != nil {
			err = uErr
		} else {
			result = &UnitBatchAndSerialInfoRsp{Info: info}
		}

	case TypeGetInfo:
		result = &GetInfoRsp{Info: s.ctrl.GetInfo()}

	case TypeGetScopeModel:
		result = &GetScopeModelRsp{Model: s.ctrl.GetScopeModel()}

	case TypeGetMaxSamplingRate:
		result = &GetMaxSamplingRateRsp{Rate: s.ctrl.GetMaxSamplingRate()}

	case TypeGetStreamEnabled:
		result = &GetStreamEnabledRsp{Enabled: s.ctrl.GetStreamEnabled()}

	case TypeGetSamplingTimeInterval:
		result = &GetSamplingTimeIntervalRsp{Interval: s.ctrl.GetSamplingTimeInterval()}

	case TypeGetTimeBase:
		result = &GetTimeBaseRsp{TimeBase: s.ctrl.GetTimeBase()}

	case TypeNumberOfEnabledAnalogChannels:
		result = &NumberOfEnabledAnalogChannelsRsp{Count: s.ctrl.NumberOfEnabledAnalogChannels()}

	case TypeGetEtsLimits:
		maxI, maxC := s.ctrl.GetEtsLimits()
		result = &GetEtsLimitsRsp{MaxInterleave: maxI, MaxCycles: maxC}

	case TypeIsDemo:
		result = &IsDemoRsp{IsDemo: s.ctrl.IsDemo()}

	default:
		err = fmt.Errorf("unknown command type: %s", cmd.Type)
	}

	if cmd.ID > 0 {
		rsp := &ResponseMessage{
			ID:      cmd.ID,
			Success: err == nil,
			Result:  result,
		}
		if err != nil {
			rsp.Error = err.Error()
		}
		_ = session.send(&ServerMessage{
			Type:    TypeResponse,
			Payload: rsp,
		})
	} else if err != nil {
		slog.Error("error executing command", "type", cmd.Type, "err", err)
	}
}
