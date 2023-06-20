package inter

const (
	Ac = iota
	Dc
	Invert
	Gen
	NsDiv
	UsDiv
	MsDiv
	SDiv
	Millivolt
	Volt
	Psec
	Nsec
	Usec
	Msec
	Sec
	Hour
	Ets
)
const (
	Eng = iota
)

var (
	vocab = [][]string{{"AC", "DC", "Invert", "Gen", "ns/div", "µs/div", "ms/div", "s/div",
		"mV", "V", "ps", "ns", "µs", "ms", "s", "h", "ETS"}}
)

func Str(language, id int) string {
	return vocab[language][id]

}
