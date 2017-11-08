package debug

import (
	"bytes"
	"fmt"
	"runtime"
)

func Stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
	return buf
}

func StackSkip(skip int) []byte {
	buf := &bytes.Buffer{}
	for ; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}

		f := runtime.FuncForPC(pc)
		fmt.Fprintf(buf, "%s(0x%x)\n", f.Name(), pc)
		fmt.Fprintf(buf, "    %s:%d\n", file, line)
	}
	return buf.Bytes()
}
