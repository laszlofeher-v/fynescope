package genericps

import (
	"fmt"
	"log/slog"
)

const (
	SimId = "sim"
)

type (
	DeviceInfo struct {
		Id          string // Handler ID (e.g., "ps2000a", "sim")
		Serial      string // Serial number (empty for simulator)
		IsSimulator bool
	}
)

type (
	ScopeHandler struct {
		EnumerateUnits   func(bufferLen int16) (count int16, serials string, serialLth int16, err error)
		OpenUnit         func(serial string) (handle int16, err error)
		OpenUnitAsync    func(serial string) (status int16, err error)
		OpenUnitProgress func() (retHandle int16, progressPercent, complete int16, err error)
		Dispatch         func(msg Message)
		Id               string
	}
)

var (
	implementedScopeHandlers []ScopeHandler
)

func Register(handler ScopeHandler) {
	implementedScopeHandlers = append(implementedScopeHandlers, handler)
}

func UnRegister(id string) {
	for i := range implementedScopeHandlers {
		if implementedScopeHandlers[i].Id == id {
			implementedScopeHandlers = append(implementedScopeHandlers[:i],
				implementedScopeHandlers[i+1:]...)
			return
		}
	}
}

func Open(serial string) (handle int16, err error) {
	for i := range implementedScopeHandlers {
		if implementedScopeHandlers[i].Id == serial {
			handle, err = implementedScopeHandlers[i].OpenUnit(serial)
			return
		}
	}
	return 0, fmt.Errorf("Scope not found")
}

func proxy(dispatch func(msg Message), con *Connection) {
	for {
		msg, ok := <-con.MsgCh // receive command from the application
		if !ok {
			slog.Error("Unexpected close of channel CmdCh")
			return
		}
		dispatch(msg) // send command to the scope and send response to the application
	}
}

func OpenSimulator(con *Connection, id string) (handle int16, err error) {
	for i := range implementedScopeHandlers {
		if implementedScopeHandlers[i].Id == id {
			handle, err = implementedScopeHandlers[i].OpenUnit("")
			go proxy(implementedScopeHandlers[i].Dispatch, con)
			return
		}
	}
	return 0, fmt.Errorf("Simulator not found")
}
func OpenUnit(con *Connection, id, serial string) (handle int16, err error) {
	for i := range implementedScopeHandlers {
		slog.Debug("OpenUnit", "implementedScopeHandlers", implementedScopeHandlers)
		if implementedScopeHandlers[i].Id == id {
			handle, err = implementedScopeHandlers[i].OpenUnit("")
			go proxy(implementedScopeHandlers[i].Dispatch, con)
			return
		}
	}
	return 0, fmt.Errorf("Simulator not found")
}

func EnumerateUnits(bufferLen int16) (count int16, serials string, serialLth int16, err error) {
	for i := range implementedScopeHandlers {
		if implementedScopeHandlers[i].Id == SimId {
			count, serials, serialLth, err = implementedScopeHandlers[i].EnumerateUnits(bufferLen)
			if err == nil {
				return
			}
		}
	}
	return
}

// EnumerateAllDevices enumerates all available devices including hardware and simulators
func EnumerateAllDevices(bufferLen int16) (devices []DeviceInfo, err error) {
	devices = make([]DeviceInfo, 0)

	// Enumerate all registered handlers
	for i := range implementedScopeHandlers {
		handler := implementedScopeHandlers[i]

		if handler.EnumerateUnits == nil {
			continue
		}

		count, serials, _, enumErr := handler.EnumerateUnits(bufferLen)
		if enumErr != nil {
			slog.Warn("EnumerateAllDevices", "handler", handler.Id, "error", enumErr)
			continue
		}

		if count > 0 {
			// Parse the serials string (comma-separated)
			serialList := parseSerials(serials, int(count))
			for _, serial := range serialList {
				devices = append(devices, DeviceInfo{
					Id:          handler.Id,
					Serial:      serial,
					IsSimulator: (handler.Id == SimId),
				})
			}
		}
	}

	if len(devices) == 0 {
		err = fmt.Errorf("no devices found")
	}

	return
}

// parseSerials splits comma-separated serial numbers
func parseSerials(serials string, count int) []string {
	if serials == "" || count == 0 {
		return []string{}
	}

	result := make([]string, 0, count)
	start := 0

	for i := 0; i < len(serials); i++ {
		if serials[i] == ',' || i == len(serials)-1 {
			end := i
			if i == len(serials)-1 && serials[i] != ',' {
				end = i + 1
			}
			if end > start {
				serial := serials[start:end]
				if serial != "" {
					result = append(result, serial)
					if len(result) >= count {
						break
					}
				}
			}
			start = i + 1
		}
	}

	return result
}
