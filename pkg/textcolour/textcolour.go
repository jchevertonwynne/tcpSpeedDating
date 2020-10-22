package textcolour

func Red(msg string) string {
	return string(formatColour(msg, []byte("\u001b[31m")))
}

func Magenta(msg string) string {
	return string(formatColour(msg, []byte("\u001b[35m")))
}

func Blue(msg string) string {
	return string(formatColour(msg, []byte("\u001b[34m")))
}

func Green(msg string) string {
	return string(formatColour(msg, []byte("\u001b[32m")))
}

func formatColour(msg string, colour []byte) []byte {
	return append(append(colour, msg...), []byte("\u001b[0m")...)
}
