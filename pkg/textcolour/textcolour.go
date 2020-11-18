package textcolour

import (
	"bytes"
	"fmt"
)

func Red(f string, args ...interface{}) string {
	return string(formatColour(fmt.Sprintf(f, args...), []byte("\u001b[31m")))
}

func Magenta(f string, args ...interface{}) string {
	return string(formatColour(fmt.Sprintf(f, args...), []byte("\u001b[35m")))
}

func Blue(f string, args ...interface{}) string {
	return string(formatColour(fmt.Sprintf(f, args...), []byte("\u001b[34m")))
}

func Green(f string, args ...interface{}) string {
	return string(formatColour(fmt.Sprintf(f, args...), []byte("\u001b[32m")))
}

func formatColour(msg string, colour []byte) []byte {
	buf := new(bytes.Buffer)
	buf.Write(colour)
	buf.WriteString(msg)
	buf.WriteString("\u001b[0m")
	return buf.Bytes()
}
