# aliyun-acm

aliyun-acm是对[阿里云应用配置管理](https://help.aliyun.com/product/59604.html)的封装

[![GoDoc](https://godoc.org/github.com/xiaojiaoyu100/aliyun-acm?status.svg)](https://godoc.org/github.com/xiaojiaoyu100/aliyun-acm)

## Usage

```go
package main

import (
	"fmt"
	aliacm "github.com/xiaojiaoyu100/aliyun-acm/v2"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/config"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/info"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/observer"
)

func handle(coll map[info.Info]*config.Config) {
    i := info.Info{DataID:"YourGroup", Group: "YourDataID"}
	configI, ok := coll[i]
    if !ok {
        return 
    }   
    
    a := info.Info{DataID:"YourAnotherGroup", Group:"YourAnotherDataID"}
    configA, ok := coll[a]
    if !ok {
        return 
    }   
}

func main() {
	d, err := aliacm.New(
    		aliacm.WithAcm("addr", "tenant", "accessKey", "secretKey"),
    		// aliacm.WithKms("regionId", "accessKey", "secretKey"),
    	)
	if err != nil {
		fmt.Println(err)
		return
	}

	o1, err := observer.New(
		observer.WithInfo(
			info.Info{Group: "YourGroup", DataID: "YourDataID"},
			),
		observer.WithHandler(handle))
	if err != nil {
		return
	}
	o2, err := observer.New(
		observer.WithInfo(
			info.Info{Group: "YourGroup", DataID: "YourDataID"},
			info.Info{Group: "YourAnotherGroup", DataID: "YourAnotherDataID"},
			),
		observer.WithHandler(handle))
	if err != nil {
		return
	}

	var f = func(err error) {
		fmt.Println(err)
	}
	d.SetHook(f)

	d.Register(o1, o2)
        d.NotifyAll()

	select{}
}
```
