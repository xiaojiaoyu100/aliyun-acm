package config

// Config 返回配置
type Config struct {
	Content []byte
	ContentMD5 string
	Pulled bool
}
