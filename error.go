package aliacm

import "context"

// Error ACM错误
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	serviceUnavailableErr = Error("ServiceUnavailable")
	internalServerErr     = Error("InternalServerError")
)

// ShouldIgnore 忽略一些不想关心的错误
func ShouldIgnore(err error) bool {
	if err == serviceUnavailableErr ||
		err == internalServerErr ||
		err == context.Canceled {
		return true
	}
	return false
}
