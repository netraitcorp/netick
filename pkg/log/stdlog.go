package log

import (
	"fmt"
	std "log"
)

func StdInfo(format string, args ...interface{}) {
	format = fmt.Sprintf("[INFO] %s\n", format)
	std.Printf(format, args...)
}
