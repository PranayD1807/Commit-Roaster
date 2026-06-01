package airoaster

import (
	"strings"
)

// ErrorKind categorizes an AI provider error so callers can show a useful message.
type ErrorKind int

const (
	ErrUnknown      ErrorKind = iota
	ErrInvalidKey             // 401 / unauthenticated — user needs to fix their key
	ErrRateLimited            // 429 / resource exhausted — temporary, back off
	ErrUnavailable            // 503 / high demand — temporary, try later
	ErrTimeout                // context deadline exceeded
)

// ClassifyError inspects an error returned by RoastCommit and returns a
// human-friendly message and the kind of error so callers can react accordingly.
func ClassifyError(err error) (kind ErrorKind, msg string) {
	if err == nil {
		return ErrUnknown, ""
	}

	s := strings.ToLower(err.Error())

	switch {
	case strings.Contains(s, "deadline exceeded") || strings.Contains(s, "context deadline"):
		return ErrTimeout, "request timed out after 15s"

	case strings.Contains(s, "401") ||
		strings.Contains(s, "unauthenticated") ||
		strings.Contains(s, "invalid api key") ||
		strings.Contains(s, "api_key_invalid") ||
		strings.Contains(s, "permission_denied"):
		return ErrInvalidKey, "invalid API key — run 'commit-roaster ai enable' to reconfigure"

	case strings.Contains(s, "429") ||
		strings.Contains(s, "resource_exhausted") ||
		strings.Contains(s, "quota") ||
		strings.Contains(s, "rate limit"):
		return ErrRateLimited, "rate limit reached — too many requests, try again shortly"

	case strings.Contains(s, "503") ||
		strings.Contains(s, "unavailable") ||
		strings.Contains(s, "high demand"):
		return ErrUnavailable, "AI service temporarily unavailable — try again later"

	default:
		return ErrUnknown, err.Error()
	}
}
