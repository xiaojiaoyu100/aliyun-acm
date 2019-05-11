# aliyun-acm

aliyun-acm是对[阿里云应用配置管理](https://help.aliyun.com/product/59604.html)的封装

## Usage

```go
package main

import (
	"fmt"
	"github.com/xiaojiaoyu100/aliyun-acm"
)


func Handle(config aliacm.Config)  {
	fmt.Println(string(config.Content))
}

func main() {
	d, err := aliacm.New(
		"your_addr",
		"your_tenant",
		"your_access_key",
		"your_secret_key")
	if err != nil {
		return
	}
	var f = func(h aliacm.Unit, err error) {
		fmt.Println(err)
	}
	d.SetHook(f)
	unit := aliacm.Unit{
		Group: "your_group",
		DataID: "your_data_id",
		FetchOnce: true, // 有且仅拉取一次
		OnChange: Handle,
	}
	d.Add(unit)
	select{}
}
```
