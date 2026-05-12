package debug

import (
	"io"
	"log"
	"os"
)

var logger = log.New(io.Discard, "[DEBUG] ", log.Lmicroseconds)

func Enable() {
	logger = log.New(os.Stderr, "[DEBUG] ", log.Lmicroseconds)
}

func Printf(format string, v ...any) {
	logger.Printf(format, v...)
}

func Println(v ...any) {
	logger.Println(v...)
}
