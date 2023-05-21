package goutil

import "unsafe"

// Str2Bytes converts a string to a byte slice without making a copy of the original string data.
// This is achieved by reinterpreting the string header to a slice header.
// Caution: this function uses the unsafe package and may cause unexpected errors.
func Str2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// Bytes2Str converts a byte slice to a string without making a copy of the original slice data.
// This is achieved by reinterpreting the slice header to a string header.
// Caution: this function uses the unsafe package and may cause unexpected errors.
func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
