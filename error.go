package aliacm

import (
	"errors"
)

// Error ACM错误
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	noChangeErr = Error("NoChangeError")
)

// ShouldIgnore 忽略一些不想关心的错误
func shouldIgnore(err error) bool {
	if err == nil ||
		errors.Is(err, noChangeErr) {
		return true
	}
	return false
}
