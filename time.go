package aliacm

import (
	"time"
)

const (
	apiTimeout = 30 * time.Second
)

func timeInMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
