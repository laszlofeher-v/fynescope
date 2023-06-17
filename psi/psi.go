package psi

type (
	Picoscope interface {
		OpenUnit(serial string) (err error)
		OpenUnitAsync(serial string) (err error)
		OpenUnitProgress() (retHandle, progressPercent, complete int16, err error)
		CloseUnit() (err error)
		FlashLed(start int16) (err error)
		PingUnit() (err error)
	}
)
