package utils

import "regexp"

func IsAlpha(c byte) bool {
	return regexp.MustCompile(`^[a-zA-Z_]+$`).MatchString(string(c))
}

func IsNumeric(c byte) bool {
	return regexp.MustCompile(`^[0-9]+$`).MatchString(string(c))
}

func IsSkipable(c byte) bool {
	return c == ' ' || c == '\n' || c == '\t' || c == '\r'
}
