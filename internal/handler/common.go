package handler

import "strconv"

func ParseIntDefault(s string, defaultValue int) int {
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return parsed
}


