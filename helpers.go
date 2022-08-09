package main

var (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

func boolToInt(truth bool) int {
	if truth {
		return 1
	} else {
		return 0
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
