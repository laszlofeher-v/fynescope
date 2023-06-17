package ps2000sim

import (
	"psgo/ps2000a"
)

type (
	Ps2000sim interface {
		ps2000a.Picoscope2000Handler
	}
)
