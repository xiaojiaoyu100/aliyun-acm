package aliacm

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/xiaojiaoyu100/aliyun-acm/v2/config"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/info"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/observer"
	"github.com/xiaojiaoyu100/curlew"
	"github.com/xiaojiaoyu100/roc"

	"context"

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

// Option 参数设置
type Option struct {
	addr      string
	tenant    string
	accessKey string
	secretKey string
}

// Diamond 提供了操作阿里云ACM的能力
type Diamond struct {
	option     Option
	c          *cast.Cast
	errHook    Hook
	r          *rand.Rand
	infoColl   map[info.Info][]*observer.Observer
	filter     map[*observer.Observer]struct{}
	all        map[info.Info]*config.Config
	dispatcher *curlew.Dispatcher
	cache      *roc.Cache
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

	var monitor = func(err error) {

	}

	dispatcher, err := curlew.New(
		curlew.WithMaxWorkerNum(100),
		curlew.WithMonitor(monitor))
	if err != nil {
		return nil, err
	}

	cache, err := roc.New()
	if err != nil {
		return nil, err
	}

	d := &Diamond{
		option:     option,
		c:          c,
		r:          r,
		infoColl:   make(map[info.Info][]*observer.Observer),
		filter:     make(map[*observer.Observer]struct{}),
		all:        make(map[info.Info]*config.Config),
		dispatcher: dispatcher,
		cache:      cache,
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

// Register registers an observer list.
func (d *Diamond) Register(oo ...*observer.Observer) {
	for _, o := range oo {
		_, ok := d.filter[o]
		if ok {
			continue
		}
		d.filter[o] = struct{}{}
		for _, in := range o.Info() {
			d.infoColl[in] = append(d.infoColl[in], o)
		}
	}
	for _, o := range oo {
		for _, i := range o.Info() {
			_, ok := d.all[i]
			if ok {
				continue
			}
			d.all[i] = &config.Config{}
		}
	}

	for _, o := range oo {
		for _, i := range o.Info() {
			conf := d.all[i]
			if conf != nil && conf.Pulled {
				o.UpdateInfo(i, conf)
				continue
			}
			req := &GetConfigRequest{
				Tenant: d.option.tenant,
				Group:  i.Group,
				DataID: i.DataID,
			}
			b, err := d.GetConfig(req)
			if err != nil {
				d.checkErr(fmt.Errorf("DataID: %s, Group: %s, err: %+v", i.DataID, i.Group, err))
				continue
			}
			conf = &config.Config{
				Content:    b,
				ContentMD5: Md5(string(b)),
				Pulled:     true,
			}
			d.all[i] = conf
			oo, ok := d.infoColl[i]
			if !ok {
				continue
			}
			for _, o := range oo {
				o.UpdateInfo(i, conf)
			}
		}
	}
}

// NotifyAll is called after Register.
func (d *Diamond) NotifyAll() {
	d.hang()
	oo := make([]*observer.Observer, 0)
	for o := range d.filter {
		oo = append(oo, o)
	}
	d.notify(oo...)
}

func (d *Diamond) notify(oo ...*observer.Observer) {
	for _, o := range oo {
		if o.Ready() {
			j := curlew.NewJob()
			j.Arg = o
			j.Fn = func(ctx context.Context, arg interface{}) error {
				o := arg.(*observer.Observer)
				o.Handle()
				return nil
			}
			d.dispatcher.Submit(j)
		}
	}
}

func (d *Diamond) hang() {
	go func() {
		for {
			var infoParams []InfoParam
			for i, conf := range d.all {
				infoParams = append(infoParams, InfoParam{
					i,
					conf.ContentMD5,
				})
			}
			ret, err := d.LongPull(infoParams...)
			if err != nil {
				d.checkErr(fmt.Errorf("acm long pull failed: %+v", err))
				time.Sleep(time.Duration(randomIntInRange(1000, 1500)) * time.Millisecond)
				continue
			}
			for _, i := range ret {
				j := curlew.NewJob()
				j.Arg = i
				j.Fn = func(ctx context.Context, arg interface{}) error {
					i, ok := arg.(info.Info)
					if !ok {
						return nil
					}
					req := &GetConfigRequest{
						Tenant: d.option.tenant,
						DataID: i.DataID,
						Group:  i.Group,
					}
					b, err := d.GetConfig(req)
					if err != nil {
						d.checkErr(fmt.Errorf("DataID: %s, Group: %s, err: %+v", i.DataID, i.Group, err))
						return nil
					}
					newContentMD5 := Md5(string(b))
					conf := &config.Config{
						Content:    b,
						ContentMD5: newContentMD5,
						Pulled:     true,
					}
					d.all[i] = conf
					oo, ok := d.infoColl[i]
					if !ok {
						return nil
					}
					for _, o := range oo {
						o.HotUpdateInfo(i, conf)
					}
					d.notify(oo...)
					return nil
				}
				d.dispatcher.Submit(j)
			}
		}
	}()
}

// SetHook 用于提醒关键错误
func (d *Diamond) SetHook(h Hook) {
	d.errHook = h
}

func (d *Diamond) checkErr(err error) {
	if shouldIgnore(err) {
		return
	}
	if d.errHook == nil {
		return
	}
	d.errHook(err)
}

// LoadingProgress shows the current load progress roughly.
func (d *Diamond) LoadingProgress() string {
	var ret = 0
	for _, conf := range d.all {
		if conf.Pulled {
			ret++
		}
	}
	return fmt.Sprintf("%d/%d", ret, len(d.all))
}
