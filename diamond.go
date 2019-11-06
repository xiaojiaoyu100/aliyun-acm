package aliacm

import (
	"math/rand"
	"reflect"
	"time"

	"errors"

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
	Coll      []Observer
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
	option  Option
	c       *cast.Cast
	units   []Unit
	errHook Hook
	r       *rand.Rand
	coll    map[info][]Observer
}

// New 产生Diamond实例
func New(addr, tenant, accessKey, secretKey string, setters ...Setter) (*Diamond, error) {
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
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	d := &Diamond{
		option: option,
		c:      c,
		r:      r,
		coll:   make(map[info][]Observer),
	}

	for _, setter := range setters {
		if err := setter(d); err != nil {
			return nil, err
		}
	}

	return d, nil
}

func randomIntInRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func (d *Diamond) register(i info, o Observer) error {
	if reflect.TypeOf(o).Kind() != reflect.Ptr {
		return errors.New("type error")
	}
	d.coll[i] = append(d.coll[i], o)
	return nil
}

func (d *Diamond) notify(unit Unit, config Config) {
	i := info{
		Group:  unit.Group,
		DataID: unit.DataID,
	}
	for _, o := range d.coll[i] {
		o.Modify(unit, config)
	}
}

// Add 添加想要关心的配置单元
func (d *Diamond) Add(unit Unit) error {
	unit.ch = make(chan Config)
	d.units = append(d.units, unit)
	for _, o := range unit.Coll {
		i := info{
			Group:  unit.Group,
			DataID: unit.DataID,
		}
		if err := d.register(i, o); err != nil {
			return err
		}
	}
	var (
		contentMD5 string
	)
	go func() {
		for {
			time.Sleep(time.Duration(randomIntInRange(20, 100)) * time.Millisecond)
			newContentMD5, err := d.LongPull(unit, contentMD5)
			d.checkErr(unit, err)
			if contentMD5 == "" &&
				newContentMD5 != "" && unit.FetchOnce {
				return
			}
			// 防止网络较差情景下MD5被重置，重新请求配置，造成阿里云限流
			if newContentMD5 != "" {
				contentMD5 = newContentMD5
			}
		}
	}()

	go func() {
		for {
			select {
			case config := <-unit.ch:
				var err error
				config.Content, err = GbkToUtf8(config.Content)
				d.checkErr(unit, err)
				d.notify(unit, config)
				if unit.FetchOnce {
					return
				}
			}
		}
	}()
	return nil
}

// SetHook 用于提醒关键错误
func (d *Diamond) SetHook(h Hook) {
	d.errHook = h
}

func (d *Diamond) checkErr(unit Unit, err error) {
	if err == nil {
		return
	}
	if d.errHook == nil {
		return
	}
	d.errHook(unit, err)
}
