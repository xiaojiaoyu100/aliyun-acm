package aliacm

import (
	"log"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/xiaojiaoyu100/cast"
)

const (
	// PublicAddr 公网（测试）
	PublicAddr = "acm.aliyun.com"
	// HZAddr 华东 1（杭州）
	HZAddr = "addr-hz-internal.edas.aliyun.com"
	// QDAddr 华北 1（青岛）
	QDAddr = "addr-qd-internal.edas.aliyun.com"
	// SHAddr 华东 2（上海）
	SHAddr = "addr-sh-internal.edas.aliyun.com"
	// BJAddr 华北 2（北京）
	BJAddr = "addr-bj-internal.edas.aliyun.com"
	// SZAddr 华南 1（深圳）
	SZAddr = "addr-sz-internal.edas.aliyun.com"
	// HKAddr 香港
	HKAddr = "addr-hk-internal.edas.aliyuncs.com"
	// SingaporeAddr 新加坡
	SingaporeAddr = "addr-singapore-internal.edas.aliyun.com"
	// ApAddr 澳大利亚（悉尼）
	ApAddr = "addr-ap-southeast-2-internal.edas.aliyun.com"
	// USWest1Addr 美国（硅谷）
	USWest1Addr = "addr-us-west-1-internal.acm.aliyun.com"
	// USEast1Addr 美国（弗吉尼亚）
	USEast1Addr = "addr-us-east-1-internal.acm.aliyun.com"
	// ShanghaiFinance1Addr 华东 2（上海）金融云
	ShanghaiFinance1Addr = "addr-cn-shanghai-finance-1-internal.edas.aliyun.com"
)

// Unit 配置基本单位
type Unit struct {
	Group     string
	DataID    string
	FetchOnce bool
	OnChange  Handler
	ch        chan Config
}

// Option 参数设置
type Option struct {
	addr      string
	tenant    string
	accessKey string
	secretKey string
}

// Config 返回配置
type Config struct {
	Content []byte
}

// Diamond 提供了操作阿里云ACM的能力
type Diamond struct {
	option       Option
	c            *cast.Cast
	units        []Unit
	longPullHook Hook
}

// New 产生Diamond实例
func New(addr, tenant, accessKey, secretKey string) (*Diamond, error) {
	option := Option{
		addr:      addr,
		tenant:    tenant,
		accessKey: accessKey,
		secretKey: secretKey,
	}
	c, err := cast.New(
		cast.WithRetry(2),
		cast.WithHTTPClientTimeout(60*time.Second),
		cast.WithExponentialBackoffDecorrelatedJitterStrategy(
			time.Millisecond*200,
			time.Millisecond*500,
		),
		cast.WithLogLevel(logrus.WarnLevel),
	)
	if err != nil {
		return nil, err
	}
	d := &Diamond{
		option: option,
		c:      c,
	}
	return d, nil
}

// Add 添加想要关心的配置单元
func (d *Diamond) Add(unit Unit) {
	unit.ch = make(chan Config)
	d.units = append(d.units, unit)
	var (
		contentMD5 string
	)
	go func() {
		for {
			newContentMD5, err := d.LongPull(unit, contentMD5)
			if err != nil {
				if d.longPullHook != nil {
					d.longPullHook(unit, err)
				}
				log.Println("LongPullErr: ", err.Error())
			}
			if contentMD5 == "" &&
				newContentMD5 != "" && unit.FetchOnce {
				return
			}
			contentMD5 = newContentMD5
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			select {
			case config := <-unit.ch:
				unit.OnChange(config)
				if unit.FetchOnce {
					return
				}
			}
		}
	}()
}

// SetHook 用于提醒当长轮询发生错误的状况
func (d *Diamond) SetHook(h Hook) {
	d.longPullHook = h
}
