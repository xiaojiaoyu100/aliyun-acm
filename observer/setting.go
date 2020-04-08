package observer

import (
	"github.com/xiaojiaoyu100/aliyun-acm/v2/config"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/info"
)

// Setting p配置
type Setting func(o *Observer) error

// WithInfo loads info.
func WithInfo(ii ...info.Info) Setting {
	return func(o *Observer) error {
		for _, i := range ii {
			o.coll[i] = &config.Config{}
		}
		return nil
	}
}

// WithHandler initializes a handler.
func WithHandler(h Handler) Setting {
	return func(o *Observer) error {
		o.h = h
		return nil
	}
}
