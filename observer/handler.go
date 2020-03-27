package observer

import (
	"github.com/xiaojiaoyu100/aliyun-acm/v2/config"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/info"
)

type Handler func(coll map[info.Info]*config.Config)
