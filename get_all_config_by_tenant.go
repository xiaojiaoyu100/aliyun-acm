package aliacm

import (
	"context"
	"errors"
	"net/http"
)

// GetAllConfigByTenantRequest 获取空间配置
type GetAllConfigByTenantRequest struct {
	Tenant   string `url:"tenant"`
	PageNo   int    `url:"pageNo"`
	PageSize int    `url:"pageSize"`
}

// GetAllConfigByTenantResponse 获取空间配置回复
type GetAllConfigByTenantResponse struct {
	TotalCount     int         `json:"totalCount"`     // 总配置数
	PageNumber     int         `json:"pageNumber"`     // 分页页号
	PagesAvailable int         `json:"pagesAvailable"` // 可用分页数
	PageItems      []*PageItem `json:"pageItems"`      // 配置元素
}

// PageItem 每一项
type PageItem struct {
	AppName string `json:"appName"`
	DataID  string `json:"dataId"`
	Group   string `json:"group"`
}

// GetAllConfigByTenant 请求空间所有配置
func (d *Diamond) GetAllConfigByTenant(args *GetAllConfigByTenantRequest) (*GetAllConfigByTenantResponse, error) {
	ip, err := d.QueryIP()
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	if err := d.withUsual(args.Tenant, "")(header); err != nil {
		return nil, err
	}
	request := d.c.NewRequest().
		WithTimeout(apiTimeout).
		WithPath(acmAllConfig.String(ip)).
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
	resp := new(GetAllConfigByTenantResponse)
	err = response.DecodeFromJSON(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
