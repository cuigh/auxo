package cast

import "unsafe"

// BytesToString converts []byte to string with zero-copy
func BytesToString(buf []byte) string {
	return *(*string)(unsafe.Pointer(&buf))
}

// StringToBytes converts string to []byte with zero-copy
func StringToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
