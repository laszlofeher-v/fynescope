package control

import "strings"

type ScopeType int

const (
	ScopeUnknown ScopeType = iota
	Scope2204A
	Scope2205A
	Scope2205A_MSO
	Scope2206B
	Scope2206B_MSO
	Scope2207B
	Scope2207B_MSO
	Scope2208B
	Scope2208B_MSO
	Scope2405A
	Scope2406B
	Scope2407B
	Scope2408B
	Scope3203D
	Scope3203D_MSO
	Scope3204D
	Scope3204D_MSO
	Scope3205D
	Scope3205D_MSO
	Scope3206D
	Scope3206D_MSO
	Scope3403D
	Scope3403D_MSO
	Scope3404D
	Scope3404D_MSO
	Scope3405D
	Scope3405D_MSO
	Scope3406D
	Scope3406D_MSO
	Scope4262
	Scope4444
	Scope4823
	Scope4224A
	Scope4424A
	Scope4824A
	Scope4225A
	Scope4425A
	Scope5242D
	Scope5242D_MSO
	Scope5243D
	Scope5243D_MSO
	Scope5244D
	Scope5244D_MSO
	Scope5442D
	Scope5442D_MSO
	Scope5443D
	Scope5443D_MSO
	Scope5444D
	Scope5444D_MSO
	Scope6403E
	Scope6404E
	Scope6405E
	Scope6406E
	Scope6424E
	Scope6425E
	Scope6426E
	Scope6428E_D
	Scope6804E
	Scope6824E
)

const ScopeSimulatedOffset ScopeType = 1000

func StringToScopeType(s string) ScopeType {
	isSim := false

	if strings.HasSuffix(s, "SIM") {
		isSim = true
		s = strings.TrimSuffix(s, "SIM")
		if strings.HasSuffix(s, "MSO") {
			s = strings.TrimSuffix(s, "MSO")
			if !strings.HasSuffix(s, "_") && !strings.HasSuffix(s, " ") && !strings.HasSuffix(s, "-") {
				s += "_"
			}
			s += "MSO"
		}
	} else if strings.HasSuffix(s, "DEMO") {
		isSim = true
		s = strings.TrimSuffix(s, "DEMO")
	}

	s = strings.ReplaceAll(s, " MSO", "_MSO")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	base := parseBaseScopeType(s)

	// Fallback for legacy demo/sim names like "2407" or "2207" which are missing the "B"
	if isSim && base == ScopeUnknown {
		base = parseBaseScopeType(s + "B")
	}

	if isSim && base != ScopeUnknown {
		return base + ScopeSimulatedOffset
	}
	return base
}

func parseBaseScopeType(s string) ScopeType {
	s = strings.ReplaceAll(s, " MSO", "_MSO")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	switch s {
	case "2204A":
		return Scope2204A
	case "2205A":
		return Scope2205A
	case "2205A_MSO":
		return Scope2205A_MSO
	case "2206B":
		return Scope2206B
	case "2206B_MSO":
		return Scope2206B_MSO
	case "2207B":
		return Scope2207B
	case "2207B_MSO":
		return Scope2207B_MSO
	case "2208B":
		return Scope2208B
	case "2208B_MSO":
		return Scope2208B_MSO
	case "2405A":
		return Scope2405A
	case "2406B":
		return Scope2406B
	case "2407B":
		return Scope2407B
	case "2408B":
		return Scope2408B
	case "3203D":
		return Scope3203D
	case "3203D_MSO":
		return Scope3203D_MSO
	case "3204D":
		return Scope3204D
	case "3204D_MSO":
		return Scope3204D_MSO
	case "3205D":
		return Scope3205D
	case "3205D_MSO":
		return Scope3205D_MSO
	case "3206D":
		return Scope3206D
	case "3206D_MSO":
		return Scope3206D_MSO
	case "3403D":
		return Scope3403D
	case "3403D_MSO":
		return Scope3403D_MSO
	case "3404D":
		return Scope3404D
	case "3404D_MSO":
		return Scope3404D_MSO
	case "3405D":
		return Scope3405D
	case "3405D_MSO":
		return Scope3405D_MSO
	case "3406D":
		return Scope3406D
	case "3406D_MSO":
		return Scope3406D_MSO
	case "4262":
		return Scope4262
	case "4444":
		return Scope4444
	case "4823":
		return Scope4823
	case "4224A":
		return Scope4224A
	case "4424A":
		return Scope4424A
	case "4824A":
		return Scope4824A
	case "4225A":
		return Scope4225A
	case "4425A":
		return Scope4425A
	case "5242D":
		return Scope5242D
	case "5242D_MSO":
		return Scope5242D_MSO
	case "5243D":
		return Scope5243D
	case "5243D_MSO":
		return Scope5243D_MSO
	case "5244D":
		return Scope5244D
	case "5244D_MSO":
		return Scope5244D_MSO
	case "5442D":
		return Scope5442D
	case "5442D_MSO":
		return Scope5442D_MSO
	case "5443D":
		return Scope5443D
	case "5443D_MSO":
		return Scope5443D_MSO
	case "5444D":
		return Scope5444D
	case "5444D_MSO":
		return Scope5444D_MSO
	case "6403E":
		return Scope6403E
	case "6404E":
		return Scope6404E
	case "6405E":
		return Scope6405E
	case "6406E":
		return Scope6406E
	case "6424E":
		return Scope6424E
	case "6425E":
		return Scope6425E
	case "6426E":
		return Scope6426E
	case "6428E_D":
		return Scope6428E_D
	case "6804E":
		return Scope6804E
	case "6824E":
		return Scope6824E

	default:
		return ScopeUnknown
	}
}

func (t ScopeType) String() string {
	isSim := false
	if t >= ScopeSimulatedOffset {
		isSim = true
		t -= ScopeSimulatedOffset
	}

	str := formatBaseScopeType(t)

	if isSim && str != "Unknown" {
		if strings.HasSuffix(str, " MSO") {
			str = strings.ReplaceAll(str, " MSO", "MSO") + "SIM"
		} else {
			str += "SIM"
		}
	}
	return str
}

func formatBaseScopeType(t ScopeType) string {
	switch t {
	case Scope2204A:
		return "2204A"
	case Scope2205A:
		return "2205A"
	case Scope2205A_MSO:
		return "2205A MSO"
	case Scope2206B:
		return "2206B"
	case Scope2206B_MSO:
		return "2206B MSO"
	case Scope2207B:
		return "2207B"
	case Scope2207B_MSO:
		return "2207B MSO"
	case Scope2208B:
		return "2208B"
	case Scope2208B_MSO:
		return "2208B MSO"
	case Scope2405A:
		return "2405A"
	case Scope2406B:
		return "2406B"
	case Scope2407B:
		return "2407B"
	case Scope2408B:
		return "2408B"
	case Scope3203D:
		return "3203D"
	case Scope3203D_MSO:
		return "3203D MSO"
	case Scope3204D:
		return "3204D"
	case Scope3204D_MSO:
		return "3204D MSO"
	case Scope3205D:
		return "3205D"
	case Scope3205D_MSO:
		return "3205D MSO"
	case Scope3206D:
		return "3206D"
	case Scope3206D_MSO:
		return "3206D MSO"
	case Scope3403D:
		return "3403D"
	case Scope3403D_MSO:
		return "3403D MSO"
	case Scope3404D:
		return "3404D"
	case Scope3404D_MSO:
		return "3404D MSO"
	case Scope3405D:
		return "3405D"
	case Scope3405D_MSO:
		return "3405D MSO"
	case Scope3406D:
		return "3406D"
	case Scope3406D_MSO:
		return "3406D MSO"
	case Scope4262:
		return "4262"
	case Scope4444:
		return "4444"
	case Scope4823:
		return "4823"
	case Scope4224A:
		return "4224A"
	case Scope4424A:
		return "4424A"
	case Scope4824A:
		return "4824A"
	case Scope4225A:
		return "4225A"
	case Scope4425A:
		return "4425A"
	case Scope5242D:
		return "5242D"
	case Scope5242D_MSO:
		return "5242D MSO"
	case Scope5243D:
		return "5243D"
	case Scope5243D_MSO:
		return "5243D MSO"
	case Scope5244D:
		return "5244D"
	case Scope5244D_MSO:
		return "5244D MSO"
	case Scope5442D:
		return "5442D"
	case Scope5442D_MSO:
		return "5442D MSO"
	case Scope5443D:
		return "5443D"
	case Scope5443D_MSO:
		return "5443D MSO"
	case Scope5444D:
		return "5444D"
	case Scope5444D_MSO:
		return "5444D MSO"
	case Scope6403E:
		return "6403E"
	case Scope6404E:
		return "6404E"
	case Scope6405E:
		return "6405E"
	case Scope6406E:
		return "6406E"
	case Scope6424E:
		return "6424E"
	case Scope6425E:
		return "6425E"
	case Scope6426E:
		return "6426E"
	case Scope6428E_D:
		return "6428E-D"
	case Scope6804E:
		return "6804E"
	case Scope6824E:
		return "6824E"

	default:
		return "Unknown"
	}
}

// Base returns the underlying real scope model if the type is a simulated or demo version.
// For real scopes, it returns the scope type itself.
func (t ScopeType) Base() ScopeType {
	if t >= ScopeSimulatedOffset {
		return t - ScopeSimulatedOffset
	}

	s := formatBaseScopeType(t)
	if !strings.Contains(s, "SIM") && !strings.Contains(s, "DEMO") {
		return t
	}

	s = strings.ReplaceAll(s, "SIM", "")
	if strings.HasSuffix(s, "DEMO") {
		s = strings.ReplaceAll(s, "DEMO", "B")
	}

	// Ensure MSO suffix is properly formatted for StringToScopeType
	s = strings.ReplaceAll(s, "BMSO", "B MSO")

	base := StringToScopeType(s)
	if base == ScopeUnknown {
		return t
	}
	return base
}

// IsETSCapable returns true if the scope type supports Equivalent Time Sampling (ETS).
// This includes simulated and MSO versions of ETS-capable scopes.
func (t ScopeType) IsETSCapable() bool {
	base := t.Base()
	if base == ScopeUnknown {
		return false
	}

	baseStr := base.String()

	// Disable for 4000 family
	if strings.HasPrefix(baseStr, "4") {
		return false
	}

	// Disable for non-A/non-B legacy PS2203/PS2204
	if baseStr == "2203" || baseStr == "2204" {
		return false
	}

	return true
}
