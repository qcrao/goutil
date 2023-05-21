package goutil

import (
	"testing"
)

func TestStr2BytesAndBytes2Str(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Empty string",
			input: "",
		},
		{
			name:  "ASCII string",
			input: "Hello, World!",
		},
		{
			name:  "Unicode string",
			input: "你好，世界！",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Str2Bytes
			b := Str2Bytes(tt.input)
			expectedBytes := []byte(tt.input)
			for i, v := range b {
				if v != expectedBytes[i] {
					t.Errorf("Str2Bytes - Expected %v, but got %v", expectedBytes, b)
					break
				}
			}

			// Test Bytes2Str
			s := Bytes2Str(b)
			if s != tt.input {
				t.Errorf("Bytes2Str - Expected %s, but got %s", tt.input, s)
			}
		})
	}
}
