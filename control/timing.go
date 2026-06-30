package control

import (
	"fmt"
	"log/slog"
	"math"
)

const (
	maxTimeBase     = 1024 * 1024 * 1024
	MaxSampling100M = 100000000
	MaxSampling200M = 200000000
	MaxSampling500M = 500000000
	MaxSampling1G   = 1000000000
)

// TODO this function is for 1 GS/s maximum sampling rate models only
func timeBase1G(timeIntervalNanoseconds uint32) (timeBase uint32) {
	switch {
	case timeIntervalNanoseconds == 0: //TODO this should be an error
		timeBase = 0
	case timeIntervalNanoseconds == 1:
		timeBase = 0
	case timeIntervalNanoseconds == 2:
		timeBase = 1
	case timeIntervalNanoseconds == 4:
		timeBase = 2
	default:
		tb := float64(timeIntervalNanoseconds)*float64(125)/float64(1000) + 2
		timeBase = uint32(math.Round(tb))
	}
	return
}

func timeInterval1G(timeBase uint32) (timeIntervalNanoseconds uint32) {
	switch {
	case timeBase == 0:
		timeIntervalNanoseconds = 1
	case timeBase == 1:
		timeIntervalNanoseconds = 2
	case timeBase == 2:
		timeIntervalNanoseconds = 4
	default:
		timeIntervalNanoseconds = 1000 * (timeBase - 2) / 125
	}
	return
}

func timeBase500M(timeIntervalNanoseconds uint32) (timeBase uint32) {
	switch {
	case timeIntervalNanoseconds == 0:
		timeBase = 0
	case timeIntervalNanoseconds <= 2:
		timeBase = 0
	case timeIntervalNanoseconds <= 4:
		timeBase = 1
	case timeIntervalNanoseconds <= 8:
		timeBase = 2
	default:
		tb := float64(timeIntervalNanoseconds)*float64(625)/float64(10000) + 2
		timeBase = uint32(math.Round(tb))
	}
	return
}

func timeInterval500M(timeBase uint32) (timeIntervalNanoseconds uint32) {
	switch {
	case timeBase == 0:
		timeIntervalNanoseconds = 2
	case timeBase == 1:
		timeIntervalNanoseconds = 4
	case timeBase == 2:
		timeIntervalNanoseconds = 8
	default:
		timeIntervalNanoseconds = 10000 * (timeBase - 2) / 625
	}
	return
}

func timeBase200M(timeIntervalNanoseconds uint32) (timeBase uint32) {
	switch {
	case timeIntervalNanoseconds == 0:
		timeBase = 0
	case timeIntervalNanoseconds <= 5:
		timeBase = 0
	case timeIntervalNanoseconds <= 10:
		timeBase = 1
	case timeIntervalNanoseconds <= 20:
		timeBase = 2
	default:
		tb := float64(timeIntervalNanoseconds)*float64(250)/float64(10000) + 2
		timeBase = uint32(math.Round(tb))
	}
	return
}

func timeInterval200M(timeBase uint32) (timeIntervalNanoseconds uint32) {
	switch {
	case timeBase == 0:
		timeIntervalNanoseconds = 5
	case timeBase == 1:
		timeIntervalNanoseconds = 10
	case timeBase == 2:
		timeIntervalNanoseconds = 20
	default:
		timeIntervalNanoseconds = 10000 * (timeBase - 2) / 250
	}
	return
}

func timeBase100M(timeIntervalNanoseconds uint32) (timeBase uint32) {
	switch {
	case timeIntervalNanoseconds == 0:
		timeBase = 0
	case timeIntervalNanoseconds <= 10:
		timeBase = 0
	case timeIntervalNanoseconds <= 20:
		timeBase = 1
	case timeIntervalNanoseconds <= 40:
		timeBase = 2
	default:
		tb := float64(timeIntervalNanoseconds)*float64(125)/float64(10000) + 2
		timeBase = uint32(math.Round(tb))
	}
	return
}

func timeInterval100M(timeBase uint32) (timeIntervalNanoseconds uint32) {
	switch {
	case timeBase == 0:
		timeIntervalNanoseconds = 10
	case timeBase == 1:
		timeIntervalNanoseconds = 20
	case timeBase == 2:
		timeIntervalNanoseconds = 40
	default:
		timeIntervalNanoseconds = 10000 * (timeBase - 2) / 125
	}
	return
}

func (psControl *PscDesc) getTimeBase(sampleCount int32) (maxSamples int32, timeIntervalNanoseconds float32, err error) {
	initialTimeBase := psControl.timeBase
	for psControl.timeBase < maxTimeBase {
		timeIntervalNanoseconds, maxSamples, err = psControl.Con.GetTimebase2(psControl.timeBase, sampleCount, psControl.overSample, 0)
		if err != nil {
			slog.Debug("setTimeBase:", "err", err)
			//try next
		} else {
			return
		}
		psControl.timeBase++
		if psControl.timeBase-initialTimeBase > 10000 {
			err = fmt.Errorf("failed to find valid timebase after 10000 attempts, last error: %v", err)
			return
		}
	}
	return
}
