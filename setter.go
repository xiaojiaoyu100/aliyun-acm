package aliacm

import "github.com/aliyun/alibaba-cloud-sdk-go/services/kms"

// Setter configures the diamond.
type Setter func(d *Diamond) error

// WithAcm 初始化Acm配置
func WithAcm(addr, tenant, accessKey, secretKey string) Setter {
	return func(d *Diamond) error {
		d.option.addr = addr
		d.option.tenant = tenant
		d.option.accessKey = accessKey
		d.option.secretKey = secretKey
		return nil
	}
}

// WithKms 初始化Kms Client
func WithKms(regionID, accessKey, secretKey string) Setter {
	return func(d *Diamond) error {
		kmsClient, err := kms.NewClientWithAccessKey(regionID, accessKey, secretKey)
		if err != nil {
			return err
		}
		d.kmsClient = kmsClient
		return nil
	}
}
