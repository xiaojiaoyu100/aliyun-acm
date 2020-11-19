package aliacm

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/xiaojiaoyu100/cast"
	"github.com/xiaojiaoyu100/lizard/convert"
)

// GetConfigRequest 获取配置参数
type GetConfigRequest struct {
	Tenant string `url:"tenant"`
	DataID string `url:"dataId"`
	Group  string `url:"group"`
}

type GetConfigResponse struct {
	Content        []byte
	DecryptContent []byte
}

// GetConfig 获取配置
func (d *Diamond) GetConfig(args *GetConfigRequest) (*GetConfigResponse, error) {
	if len(args.Group) == 0 {
		args.Group = DefaultGroup
	}
	if len(args.Tenant) == 0 {
		args.Tenant = d.option.tenant
	}
	ip, err := d.QueryIP()
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	if err := d.withSignature(args.Tenant, args.Group)(header); err != nil {
		return nil, err
	}
	request := d.c.NewRequest().
		WithTimeout(apiTimeout).
		WithPath(acmConfig.String(ip)).
		WithQueryParam(args).
		WithHeader(header).
		Get()
	response, err := d.c.Do(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	if !response.Success() {
		return nil, errors.New(response.String())
	}

	config, err := d.getConfig(response, args.DataID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// getConfig 适配配置kms加密
func (d *Diamond) getConfig(response *cast.Response, dataID string) (*GetConfigResponse, error) {
	config := &GetConfigResponse{
		Content:        response.Body(),
		DecryptContent: response.Body(),
	}
	if d.kmsClient == nil {
		return config, nil
	}

	body := convert.ByteToString(response.Body())
	switch {
	case strings.HasPrefix(dataID, "cipher-kms-aes-128-"):
		dataKey, err := d.kmsDecrypt(response.Header().Get("Encrypted-Data-Key"))
		if err != nil {
			return nil, err
		}

		bodyByte, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			return nil, err
		}
		dataKeyByte, err := base64.StdEncoding.DecodeString(dataKey)
		if err != nil {
			return nil, err
		}

		config.DecryptContent, err = aesDecrypt(bodyByte, dataKeyByte)
		if err != nil {
			return nil, err
		}

	case strings.HasPrefix(dataID, "cipher-"):
		configStr, err := d.kmsDecrypt(body)
		if err != nil {
			return nil, err
		}
		config.DecryptContent = convert.StringToByte(configStr)

	}

	return config, nil
}
