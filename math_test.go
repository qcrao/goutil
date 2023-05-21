package goutil

import "testing"

func TestFloat64Equal(t *testing.T) {
	tests := []struct {
		f1        float64
		f2        float64
		precision uint8
		want      bool
	}{
		{0.12345677, 0.12345679, 8, false},
		{0.12345678, 0.12345679, 8, true},

		{0.12345678, 0.12345670, 8, false},
		{0.12345678, 0.12345678, 8, true},

		{0.12345678, 0.12345679, 9, false},
		{0.12345678, 0.12345678, 9, true},
	}

	for _, tt := range tests {
		if got := Float64Equal(tt.f1, tt.f2, tt.precision); got != tt.want {
			t.Errorf("Float64Equal(%v, %v, %v) = %v, want %v", tt.f1, tt.f2, tt.precision, got, tt.want)
		}
	}
}
