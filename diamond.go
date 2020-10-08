package aliacm

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/kms"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/config"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/info"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/observer"
	"github.com/xiaojiaoyu100/curlew"
	"github.com/xiaojiaoyu100/roc"

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
	kmsClient  *kms.Client
	errHook    Hook
	r          *rand.Rand
	infoColl   sync.Map // info.Info: []*observer.Observer
	filter     sync.Map // *observer.Observer: struct{｝
	all        sync.Map // info.Info: *config.Config
	dispatcher *curlew.Dispatcher
	cache      *roc.Cache
}

// New 产生Diamond实例
func New(setters ...Setter) (*Diamond, error) {
	if len(setters) == 0 {
		return nil, errors.New("Lack of setter ")
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
		c:          c,
		r:          r,
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
		_, ok := d.filter.Load(o)
		if ok {
			continue
		}
		d.filter.Store(o, struct{}{})
		for _, i := range o.Info() {
			oov, ok := d.infoColl.Load(i)
			var oo []*observer.Observer
			if ok {
				oo = oov.([]*observer.Observer)
			}
			oo = append(oo, o)
			d.infoColl.Store(i, oo)
		}
	}
	for _, o := range oo {
		for _, i := range o.Info() {
			_, ok := d.all.Load(i)
			if ok {
				continue
			}
			d.all.Store(i, &config.Config{})
		}
	}

	for _, o := range oo {
		for _, i := range o.Info() {
			confv, _ := d.all.Load(i)
			var conf *config.Config
			if confv != nil {
				conf, _ = confv.(*config.Config)
			}
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
			d.all.Store(i, conf)

			oov, ok := d.infoColl.Load(i)
			if !ok {
				continue
			}
			oo, ok := oov.([]*observer.Observer)
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
	d.filter.Range(func(key, value interface{}) bool {
		o, ok := key.(*observer.Observer)
		if !ok {
			return true
		}
		oo = append(oo, o)
		return true
	})
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
			time.Sleep(time.Duration(randomIntInRange(1000, 1500)) * time.Millisecond)
			var infoParams []InfoParam
			d.all.Range(func(key, value interface{}) bool {
				i, ok := key.(info.Info)
				if !ok {
					return true
				}
				conf, ok := value.(*config.Config)
				if !ok {
					return true
				}
				infoParams = append(infoParams, InfoParam{
					i,
					conf.ContentMD5,
				})
				return true
			})
			ret, err := d.LongPull(infoParams...)
			if err != nil {
				d.checkErr(fmt.Errorf("acm long pull failed: %+v", err))
				continue
			}
			if len(ret) == 0 {
				continue
			}
			for _, i := range ret {
				j := curlew.NewJob()
				j.Arg = i
				j.Fn = d.update
				d.dispatcher.Submit(j)
			}
		}
	}()
}

func (d *Diamond) update(ctx context.Context, arg interface{}) error {
	i, ok := arg.(info.Info)
	if !ok {
		return nil
	}
	preConfv, ok := d.all.Load(i)
	if !ok {
		return nil
	}
	preConf, ok := preConfv.(*config.Config)
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

	if preConf.ContentMD5 == newContentMD5 {
		return nil
	}

	conf := &config.Config{
		Content:    b,
		ContentMD5: newContentMD5,
		Pulled:     true,
	}
	d.all.Store(i, conf)
	oov, ok := d.infoColl.Load(i)
	if !ok {
		return nil
	}
	oo, ok := oov.([]*observer.Observer)
	if !ok {
		return nil
	}
	for _, o := range oo {
		o.HotUpdateInfo(i, conf)
	}
	d.notify(oo...)
	return nil
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
