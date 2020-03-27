package aliacm

import "github.com/xiaojiaoyu100/aliyun-acm/v2/info"

// Hook 提供了长轮询失败发生的回调
type Hook func(i info.Info, err error)
