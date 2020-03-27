package aliacm

import (
	"context"
	"errors"
	"net/http"
)

// PublishConfigRequest 发布配置请求
type PublishConfigRequest struct {
	Tenant  string `url:"tenant"`
	DataID  string `url:"dataId"`
	Group   string `url:"group"`
	Content string `url:"content"`
}

// PublishConfig 发布配置
func (d *Diamond) PublishConfig(args *PublishConfigRequest) error {
	if len(args.Group) == 0 {
		args.Group = DefaultGroup
	}
	ip, err := d.QueryIP()
	if err != nil {
		return err
	}
	header := make(http.Header)
	if err := d.withUsual(args.Tenant, args.Group)(header); err != nil {
		return err
	}
	request := d.c.NewRequest().
		WithTimeout(apiTimeout).
		WithPath(acmPublishConfig.String(ip)).
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
