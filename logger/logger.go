package logger

import (
	"fmt"
	"log"
	"strings"
)

// const (
// 	blue   = "\033[34m"
// 	red    = "\033[31m"
// 	yellow = "\033[33m"
// 	reset  = "\033[0m"
// )

func Info(message ...any) {
	msg := fmt.Sprint(message...)
	// log.Printf("\033[34m%v\033[0m\n", msg) // Removed blue
	log.Printf("\033[0m%v\033[0m\n", msg)
}
func Warning(message ...any) {
	msg := fmt.Sprint(message...)
	log.Printf("\033[33m%v\033[0m\n", msg)
}
func Error(message ...any) {
	msg := fmt.Sprint(message...)
	log.Printf("\033[31m%v\033[0m\n", msg)
}

func Normal(message ...any) {
	msg := fmt.Sprint(message...)
	log.Printf("\033[0m%v\033[0m\n", msg)
}

func Header(message ...any) {
	msg := fmt.Sprint(message...)
	line := strings.Repeat("-", len(msg)+1)
	log.Printf("\033[0m%v\033[0m\n", line)
	log.Printf("\033[0m%v\033[0m\n", msg)
	log.Printf("\033[0m%v\033[0m\n", line)
}
