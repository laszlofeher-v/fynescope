package control

import (
	"testing"
)

func TestTimeBase500M(t *testing.T) {
	tests := []struct {
		timeInterval uint32
		want         uint32
	}{
		{0, 0},
		{1, 0},
		{2, 0},
		{3, 1},
		{4, 1},
		{5, 2},
		{8, 2},
		{10, 3},
		{18, 3}, // round(18 * 625 / 10000 + 2) = round(1.125 + 2) = 3
		{26, 4}, // round(26 * 625 / 10000 + 2) = round(1.625 + 2) = 4
	}
	for _, tt := range tests {
		if got := timeBase500M(tt.timeInterval); got != tt.want {
			t.Errorf("timeBase500M(%v) = %v, want %v", tt.timeInterval, got, tt.want)
		}
	}
}

func TestTimeInterval500M(t *testing.T) {
	tests := []struct {
		timeBase uint32
		want     uint32
	}{
		{0, 2},
		{1, 4},
		{2, 8},
		{3, 16}, // 10000 * (3-2) / 625 = 10000 / 625 = 16
		{4, 32}, // 10000 * (4-2) / 625 = 20000 / 625 = 32
	}
	for _, tt := range tests {
		if got := timeInterval500M(tt.timeBase); got != tt.want {
			t.Errorf("timeInterval500M(%v) = %v, want %v", tt.timeBase, got, tt.want)
		}
	}
}

func TestTimeBase200M(t *testing.T) {
	tests := []struct {
		timeInterval uint32
		want         uint32
	}{
		{0, 0},
		{5, 0},
		{10, 1},
		{20, 2},
		{40, 3}, // round(40 * 250 / 10000 + 2) = round(1 + 2) = 3
	}
	for _, tt := range tests {
		if got := timeBase200M(tt.timeInterval); got != tt.want {
			t.Errorf("timeBase200M(%v) = %v, want %v", tt.timeInterval, got, tt.want)
		}
	}
}

func TestTimeInterval200M(t *testing.T) {
	tests := []struct {
		timeBase uint32
		want     uint32
	}{
		{0, 5},
		{1, 10},
		{2, 20},
		{3, 40}, // 10000 * (3-2) / 250 = 40
	}
	for _, tt := range tests {
		if got := timeInterval200M(tt.timeBase); got != tt.want {
			t.Errorf("timeInterval200M(%v) = %v, want %v", tt.timeBase, got, tt.want)
		}
	}
}

func TestTimeBase100M(t *testing.T) {
	tests := []struct {
		timeInterval uint32
		want         uint32
	}{
		{0, 0},
		{10, 0},
		{20, 1},
		{40, 2},
		{80, 3}, // round(80 * 125 / 10000 + 2) = round(1 + 2) = 3
	}
	for _, tt := range tests {
		if got := timeBase100M(tt.timeInterval); got != tt.want {
			t.Errorf("timeBase100M(%v) = %v, want %v", tt.timeInterval, got, tt.want)
		}
	}
}

func TestTimeInterval100M(t *testing.T) {
	tests := []struct {
		timeBase uint32
		want     uint32
	}{
		{0, 10},
		{1, 20},
		{2, 40},
		{3, 80}, // 10000 * (3-2) / 125 = 80
	}
	for _, tt := range tests {
		if got := timeInterval100M(tt.timeBase); got != tt.want {
			t.Errorf("timeInterval100M(%v) = %v, want %v", tt.timeBase, got, tt.want)
		}
	}
}
