package observer

import (
	"github.com/xiaojiaoyu100/aliyun-acm/v2/config"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/info"
)

// Observer 观察者
type Observer struct {
	coll     map[info.Info]*config.Config
	h        Handler
	consumed bool
}

// New 生成一个观察者
func New(ss ...Setting) (*Observer, error) {
	o := &Observer{}
	o.coll = make(map[info.Info]*config.Config)
	for _, s := range ss {
		if err := s(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

// Ready 是否准备好
func (o *Observer) Ready() bool {
	var ret = true
	for _, c := range o.coll {
		if !c.Pulled {
			ret = false
			break
		}
	}
	return ret
}

// Info 观察数组
func (o *Observer) Info() []info.Info {
	var ii []info.Info
	for i := range o.coll {
		ii = append(ii, i)
	}
	return ii
}

// Handle 处理函数
func (o *Observer) Handle() {
	if o.consumed {
		return
	}
	o.consumed = true
	o.h(o.coll)
}

// UpdateInfo 更新配置
func (o *Observer) UpdateInfo(i info.Info, conf *config.Config) {
	_, ok :=  o.coll[i]
	if !ok {
		return
	}
	o.coll[i] = conf
}

// HotUpdateInfo 热更新配置
func (o *Observer) HotUpdateInfo(i info.Info, conf *config.Config) {
	_, ok :=  o.coll[i]
	if !ok {
		return
	}
	o.consumed = false
	o.coll[i] = conf
}
