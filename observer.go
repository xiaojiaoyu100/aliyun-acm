package aliacm

import (
	"sync"
)



// AfterUpdateHook 配置更新完毕后的回调函数
type AfterUpdateHook func([]Config)

// Observer observes the config change.
type Observer struct {
	AfterUpdateHook AfterUpdateHook
	confs           sync.Map
	infos           []Info
}

// AddInfo 用来添加想要关心的配置
func (o *Observer) AddInfo(ufs ...Info) {
	for _, uf := range ufs {
		o.confs.LoadOrStore(uf, nil)
		o.infos = append(o.infos, uf)
	}
}

// Infos 获取Observer所有的Info
func (o *Observer) Infos() []Info {
	return o.infos[:]
}

// OnUpdate ACM配置更新后的回调函数
func (o *Observer) OnUpdate(config Config) {
	foundFlag := false
	readFlag := true
	var copyUnits []Config

	o.confs.Range(func(key, valueIf interface{}) bool {
		if flag, ok := key.(Info); ok && flag.Group == config.Group && flag.DataID == config.DataID {
			o.confs.Store(flag, config)
			foundFlag = true
			copyUnits = append(copyUnits, config)
			return true
		}

		if valueIf == nil {
			readFlag = false
			return true
		}

		if realConfig, ok := valueIf.(Config); ok {
			copyUnits = append(copyUnits, realConfig)
		}

		return true
	})

	if readFlag && foundFlag && o.AfterUpdateHook != nil {
		o.AfterUpdateHook(copyUnits)
	}
}
