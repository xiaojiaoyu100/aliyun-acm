package aliacm

import (
	"context"
	"errors"
	"github.com/xiaojiaoyu100/aliyun-acm/v2/info"
	"net/http"
	"net/url"
	"strings"
)

const (
	wordSeparator = string(rune(2))
	lineSeparator = string(rune(1))
)

// LongPull 监听配置
func (d *Diamond) LongPull(info info.Info, contentMD5 string) ([]byte, string, error) {
	ip, err := d.QueryIP()
	if err != nil {
		return nil, "", err
	}
	switch contentMD5 {
	case "":
		args := new(GetConfigRequest)
		args.Tenant = d.option.tenant
		args.Group = info.Group
		args.DataID = info.DataID
		content, err := d.GetConfig(args)
		if err != nil {
			return nil, "", err
		}
		contentMD5 = Md5(string(content))
		return content, contentMD5, nil
	default:
		headerSetters := []headerSetter{
			d.withLongPollingTimeout(),
			d.withUsual(d.option.tenant, info.Group),
		}
		header := make(http.Header)
		for _, setter := range headerSetters {
			if err := setter(header); err != nil {
				return nil, "", err
			}
		}
		var longPollRequest struct {
			ProbeModifyRequest string `url:"Probe-Modify-Request"`
		}
		longPollRequest.ProbeModifyRequest = strings.Join([]string{info.DataID, info.Group, contentMD5, d.option.tenant}, wordSeparator) + lineSeparator
		request := d.c.NewRequest().
			WithPath(acmLongPull.String(ip)).
			WithFormURLEncodedBody(longPollRequest).
			WithHeader(header).
			Post()
		response, err := d.c.Do(context.TODO(), request)
		if err != nil {
			return nil, "", err
		}
		switch response.StatusCode() {
		case http.StatusServiceUnavailable:
			return nil, "", serviceUnavailableErr
		case http.StatusInternalServerError:
			return nil, "", internalServerErr
		}
		if !response.Success() {
			return nil, "", errors.New(response.String())
		}
		ret := url.QueryEscape(strings.Join([]string{info.DataID, info.Group, d.option.tenant}, wordSeparator) + lineSeparator)
		if ret == strings.TrimSpace(response.String()) {
			args := new(GetConfigRequest)
			args.Tenant = d.option.tenant
			args.Group = info.Group
			args.DataID = info.DataID
			content, err := d.GetConfig(args)
			if err != nil {
				return nil, "", err
			}
			contentMD5 := Md5(string(content))
			return content, contentMD5, nil
		}
		return nil, "", errors.New("long pull unexpected error")
	}
}
