package gui

import (
	"log/slog"
	"fynescope/control"
	"fynescope/genericps"

	"fyne.io/fyne/v2"
)

var (
	frqUnits    = [...]string{" Hz", " kHz", " MHz"}
	periodUnits = [...]string{" ns", " µs", " ms", " s"}
)

const (
	maxFrqDisp    = 9999
	maxPeriodDisp = 9999
)

type measureState struct {
	min, max         float32
	count            int
	mean             float32
	meanOk           bool
	hysteresisTop    float32
	hysteresisBottom float32
}

type MeasureDesc struct {
	average [genericps.MaxChannel]measureState
}

func measureFrq(receiveBuffer []float32, mean, hTop, hBottom float32, timeInterval float64) (frq, period float64) {
	/*
		1
		hTop    ...................................
		2, 6
		mean    ...................................
		3, 5
		hBottom ...................................
		4
	*/
	type (
		stateDesc int
	)
	const (
		s1 stateDesc = iota
		s2
		s3
		s4
		s5
		s6
	)
	var state stateDesc
	raising := true
	v0 := float32(receiveBuffer[0])
	switch {
	case v0 >= hTop:
		state = s1
	case v0 >= mean:
		if len(receiveBuffer) < 2 {
			slog.Debug("measureFrq", "len(receiveBuffer)", len(receiveBuffer))
			return
		}
		if float32(receiveBuffer[1]) > v0 {
			state = s6
		} else {
			state = s2
			raising = false
		}
	case v0 >= hBottom:
		if len(receiveBuffer) < 2 {
			slog.Debug("measureFrq", "len(receiveBuffer)", len(receiveBuffer))
			return
		}
		if float32(receiveBuffer[1]) > v0 {
			state = s5
		} else {
			state = s3
			raising = false
		}
	default:
		state = s4
	}
	halfPeriodCount := 0
	start := 0
	end := 0
	// slog.Debug("frq", "mean", mean, "uh", uh, "lh", lh)
	for i := 1; i < len(receiveBuffer); i++ {
		v := float32(receiveBuffer[i])
		// slog.Debug("frq", "v", v, "mean", mean, "d", d)
		switch state {
		case s1:
			switch {
			case v >= hTop:
			case v >= mean:
				state = s2
			case v >= hBottom:
				state = s3
				if !raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = true
				}
			default:
				state = s4
				if !raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = true
				}
			}
		case s2:
			switch {
			case v >= hTop:
				state = s1
			case v >= mean:
				state = s2
			case v >= hBottom:
				state = s3
				if !raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = true
				}
			default:
				state = s4
				if !raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = true
				}
			}
		case s3:
			switch {
			case v >= hTop:
				state = s1
			case v >= mean:
				state = s2
			case v >= hBottom:
				state = s3
			default:
				state = s4
			}
		case s4:
			switch {
			case v >= hTop:
				state = s1
				if raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = false
				}
			case v >= mean:
				state = s6
				if raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = false
				}
			case v >= hBottom:
				state = s5
			}
		case s5:
			switch {
			case v >= hTop:
				state = s1
				if raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = false
				}
			case v >= mean:
				state = s6
				if raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = false
				}
			case v >= hBottom:
				state = s5
			default:
				state = s4
			}
		case s6:
			switch {
			case v >= hTop:
				state = s1
			case v >= mean:
				state = s6
			case v >= hBottom:
				state = s5
			default:
				state = s4
			}
		}
	}
	halfPeriodCount--
	t := end - start
	// slog.Debug("measureFrq", "start", timeInterval*float32(start), "end", timeInterval*float32(end))
	// slog.Debug("measureFrq", "t", t, "halfPeriodCount", halfPeriodCount)
	period = (float64(timeInterval) * float64(t)) / (float64(halfPeriodCount) / 2)
	// period = int(math.Round(periodf * 1e9))
	if period == 0 {
		frq = 0
	} else {
		frq = 1 / period
	}
	// slog.Debug("frq", "mean", mean, "count", count, "t", t, "frq", frq)
	// slog.Debug("frq", "mean", mean, "count", count, "t", t, "frq", frq)
	return
}

func measureFrqEts(receiveBuffer []float32, etsBuffer []int64, mean, hTop, hBottom float32) (frq, period float64) {
	/*
		1
		hTop    ...................................
		2, 6
		mean    ...................................
		3, 5
		hBottom ...................................
		4
	*/
	type (
		stateDesc int
	)
	const (
		s1 stateDesc = iota
		s2
		s3
		s4
		s5
		s6
	)
	if len(receiveBuffer) == 0 || len(etsBuffer) == 0 {
		return 0, 0
	}
	var state stateDesc
	raising := true
	v0 := float32(receiveBuffer[0])
	switch {
	case v0 >= hTop:
		state = s1
	case v0 >= mean:
		if len(receiveBuffer) < 2 {
			slog.Debug("measureFrqEts", "len(receiveBuffer)", len(receiveBuffer))
			return
		}
		if float32(receiveBuffer[1]) > v0 {
			state = s6
		} else {
			state = s2
			raising = false
		}
	case v0 >= hBottom:
		if len(receiveBuffer) < 2 {
			slog.Debug("measureFrqEts", "len(receiveBuffer)", len(receiveBuffer))
			return
		}
		if float32(receiveBuffer[1]) > v0 {
			state = s5
		} else {
			state = s3
			raising = false
		}
	default:
		state = s4
	}
	halfPeriodCount := 0
	start := 0
	end := 0
	for i := 1; i < len(receiveBuffer) && i < len(etsBuffer); i++ {
		v := float32(receiveBuffer[i])
		switch state {
		case s1:
			switch {
			case v >= hTop:
			case v >= mean:
				state = s2
			case v >= hBottom:
				state = s3
				if !raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = true
				}
			default:
				state = s4
				if !raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = true
				}
			}
		case s2:
			switch {
			case v >= hTop:
				state = s1
			case v >= mean:
				state = s2
			case v >= hBottom:
				state = s3
				if !raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = true
				}
			default:
				state = s4
				if !raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = true
				}
			}
		case s3:
			switch {
			case v >= hTop:
				state = s1
			case v >= mean:
				state = s2
			case v >= hBottom:
				state = s3
			default:
				state = s4
			}
		case s4:
			switch {
			case v >= hTop:
				state = s1
				if raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = false
				}
			case v >= mean:
				state = s6
				if raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = false
				}
			case v >= hBottom:
				state = s5
			}
		case s5:
			switch {
			case v >= hTop:
				state = s1
				if raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = false
				}
			case v >= mean:
				state = s6
				if raising {
					if halfPeriodCount == 0 {
						start = i
					} else {
						end = i
					}
					halfPeriodCount++
					raising = false
				}
			case v >= hBottom:
				state = s5
			default:
				state = s4
			}
		case s6:
			switch {
			case v >= hTop:
				state = s1
			case v >= mean:
				state = s6
			case v >= hBottom:
				state = s5
			default:
				state = s4
			}
		}
	}
	halfPeriodCount--

	if start >= len(etsBuffer) || end >= len(etsBuffer) {
		return 0, 0
	}

	t_femtoseconds := float64(etsBuffer[end] - etsBuffer[start])

	if halfPeriodCount <= 0 || t_femtoseconds <= 0 {
		return 0, 0
	}

	period = (t_femtoseconds * 1e-15) / (float64(halfPeriodCount) / 2.0)
	if period == 0 {
		frq = 0
	} else {
		frq = 1.0 / period
	}
	return
}
func (scp *ScpDesc) numOfMeasurements() (numberOfMeasurements int) {
	if scp.maxScreenTime > 0.1 {
		numberOfMeasurements = 1
	} else {
		numberOfMeasurements = 32
	}
	return
}

func (scp *ScpDesc) UpdateMeasurements(buffers [][]int16, samplingTimeInterval float64) {
	numberOfMeasurements := scp.numOfMeasurements()
	for channelIndex := range buffers {
		channel := &scp.Settings.Channels[channelIndex]
		channelViewer := &scp.channelViewers[channelIndex]
		receiveBuffer := buffers[channelIndex]
		displayBuffer := scp.displayBuffers[channelIndex]

		if channel.Enabled {
			avg := &scp.Measure.average[channelIndex]
			avg.count++
			min := float32(32767)
			max := float32(-32768)
			scale := float32(genericps.InputRanges[channel.VRange]) / float32(scp.MaxValue)

			for i := range receiveBuffer {
				displayBuffer[i] = float32(receiveBuffer[i]) * scale
			}

			scp.applyDigitalFilters(channelIndex, displayBuffer, samplingTimeInterval)

			for i := range displayBuffer {
				v := displayBuffer[i]
				if v > max {
					max = v
				}
				if v < min {
					min = v
				}
			}
			avg.max += max
			avg.min += min

			if avg.count >= numberOfMeasurements {
				nf := float32(numberOfMeasurements)
				avg.min = avg.min / nf
				avg.max = avg.max / nf
				minViewer := channelViewer.minV
				maxViewer := channelViewer.maxV
				min = avg.min
				max = avg.max
				minViewer.SetFloatValue(float64(min/1000), 3)
				maxViewer.SetFloatValue(float64(max/1000), 3)

				avg.max = 0
				avg.min = 0
				avg.count = 0
				mean := (min + max) / 2
				var f, period float64
				if scp.triggerSettingMsg.Mode == control.ETS {
					f, period = measureFrqEts(displayBuffer, scp.etsBuffer, mean, mean+0.8*(max-mean),
						mean-0.8*(mean-min))
				} else {
					f, period = measureFrq(displayBuffer, mean, mean+0.8*(max-mean),
						mean-0.8*(mean-min), samplingTimeInterval)
				}

				if f == 0 {
					scp.channelViewers[channelIndex].frq.SetFloatValue(0, 0)
					scp.channelViewers[channelIndex].period.SetFloatValue(0, 0)
					continue
				}

				dpos := 0
				unit := 0
				switch {
				case f < maxFrqDisp/1000:
					unit = 0
					dpos = 3
				case f < maxFrqDisp/100:
					unit = 0
					dpos = 2
				case f < maxFrqDisp/10:
					unit = 0
					dpos = 1
				case f < maxFrqDisp:
				case f < 10*maxFrqDisp:
					f /= 1000
					unit = 1
					dpos = 2
				case f < 100*maxFrqDisp:
					f /= 1000
					unit = 1
					dpos = 1
				case f < 1000*maxFrqDisp:
					f /= 1000000
					unit = 2
					dpos = 3
				case f < 10000*maxFrqDisp:
					f /= 1000000
					unit = 2
					dpos = 2
				case f < 100000*maxFrqDisp:
					f /= 1000000
					unit = 2
					dpos = 1
				}
				scp.channelViewers[channelIndex].frq.SetFloatValue(f, dpos)
				scp.channelViewers[channelIndex].frq.SetUnit(frqUnits[unit])
				fyne.Do(scp.channelViewers[channelIndex].frq.Refresh)

				dpos = 0
				unit = 0
				period = 1e9 * period
				switch {
				case period < maxPeriodDisp:
				case period < 10*maxPeriodDisp:
					period /= 1000
					unit = 1
					dpos = 2
				case period < 100*maxPeriodDisp:
					period /= 1000
					unit = 1
					dpos = 1
				case period < 1000*maxPeriodDisp:
					period /= 1000000
					unit = 2
					dpos = 3
				case period < 10000*maxPeriodDisp:
					period /= 1000000
					unit = 2
					dpos = 2
				case period < 100000*maxPeriodDisp:
					period /= 1000000
					unit = 2
					dpos = 1
				case period < 1000000*maxPeriodDisp:
					period /= 10000000
					unit = 3
					dpos = 1
				case period < 10000000*maxPeriodDisp:
					period /= 100000000
					unit = 3
					dpos = 1
				}
				scp.channelViewers[channelIndex].period.SetFloatValue(period, dpos)
				scp.channelViewers[channelIndex].period.SetUnit(periodUnits[unit])
				fyne.Do(scp.channelViewers[channelIndex].period.Refresh)
			}
		}
	}
	if scp.controlTab != nil && scp.controlTab.SelectedIndex() == ffTabIndex {
		scp.processFfData()
		if scp.ffRaster != nil {
			fyne.Do(scp.ffRaster.Refresh)
		}
	}
}
