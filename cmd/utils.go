package cmd

import (
	"os"
	"strconv"
)

// --- Flag helpers ---

func HasFlag(flags ...string) bool {
	for _, arg := range os.Args[2:] {
		for _, f := range flags {
			if arg == f {
				return true
			}
		}
	}
	return false
}

func GetFlagInt(def int, flags ...string) int {
	args := os.Args[2:]
	for i, arg := range args {
		for _, f := range flags {
			if arg == f && i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					return v
				}
			}
		}
	}
	return def
}

// --- ANSI color helpers ---

func Bold(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[1m" + s + "\033[0m"
}

func Dim(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[2m" + s + "\033[0m"
}

func Green(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}
