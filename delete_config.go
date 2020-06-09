package aliacm

import (
	"context"
	"errors"
	"net/http"
)

// DeleteConfigRequest 删除配置请求
type DeleteConfigRequest struct {
	Tenant string `url:"tenant"`
	DataID string `url:"dataId"`
	Group  string `url:"group"`
}

// DeleteConfig 删除配置
func (d *Diamond) DeleteConfig(args *DeleteConfigRequest) error {
	if len(args.Group) == 0 {
		args.Group = DefaultGroup
	}
	ip, err := d.QueryIP()
	if err != nil {
		return err
	}

	header := make(http.Header)

	if err := d.withSignature(args.Tenant, args.Group)(header); err != nil {
		return err
	}

	request := d.c.NewRequest().
		WithTimeout(apiTimeout).
		WithPath(acmDeleteConfig.String(ip)).
		WithFormURLEncodedBody(args).
		WithHeader(header).
		Post()
	response, err := d.c.Do(context.TODO(), request)
	if err != nil {
		return err
	}
	if !response.Success() {
		return errors.New(response.String())
	}
	return nil
}
