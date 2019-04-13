# Aliyun ACM

*maintainer: CJK <regan.cjk@gmail.com>*

## Usage

```go
var Client *acm.Client

func init() {
	// Setup once.
	var err error
	Client, err = acm.GetClient(
		"EndPoint",
		"Tenant", // Use tenant to separate deployment environment.
		"AccessKey",
		"SecretKey",
	)
	if err != nil {
		panic(err)
	}

	// Get static config.
	value, err := Client.GetConfig("DEFAULT_GROUP", "dataID")
	if err != nil {
		log.Println(err)
	}
	// Note that value has been decoded from GBK to UTF-8.
	fmt.Println(value)

	// Listen on dynamic config in goroutine
	go Client.Listen("DEFAULT_GROUP", "dataID", func(newValue string) {
		// Do something with new config value while update.
		fmt.Println(newValue)
	})
}
```
