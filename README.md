# Aliyun ACM

*maintainer: CJK regan.cjk@gmail.com*

## Usage

```go
import "gitlab.xinghuolive.com/golang/aliyun-acm"

func init() {
	acm.Setup(
		"EndPoint",
		"Tenant",
		"AccessKey",
		"SecretKey",
	)
}

func main() {
    // Get static config
    value, md5 := acm.GetConfig("DEFAULT_GROUP", "dataID")
    fmt.Println(value)

    // Listen on dynamic config in goroutine
	go acm.Listen("DEFAULT_GROUP", "dataID", func(newValue string) {
        // Do something with new config value while update.
		fmt.Println(newValue)
	})
}
```
