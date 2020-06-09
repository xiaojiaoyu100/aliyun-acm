package aliacm

import (
	"context"
	"errors"
	"net/http"
)

// GetConfigRequest 获取配置参数
type GetConfigRequest struct {
	Tenant string `url:"tenant"`
	DataID string `url:"dataId"`
	Group  string `url:"group"`
}

// GetConfig 获取配置
func (d *Diamond) GetConfig(args *GetConfigRequest) ([]byte, error) {
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
	return response.Body(), nil
}
