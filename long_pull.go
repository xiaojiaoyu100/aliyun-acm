package aliacm

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/xiaojiaoyu100/aliyun-acm/v2/info"
)

const (
	wordSeparator = string(rune(2))
	lineSeparator = string(rune(1))
)

// InfoParam long pull 参数
type InfoParam struct {
	info.Info
	ContentMD5 string
}

// LongPull 监听配置
func (d *Diamond) LongPull(infoParams ...InfoParam) ([]info.Info, error) {
	ip, err := d.QueryIP()
	if err != nil {
		return nil, err
	}

	var longPollRequest struct {
		ProbeModifyRequest string `url:"Probe-Modify-Request"`
	}
	for _, infoParam := range infoParams {
		if longPollRequest.ProbeModifyRequest != "" {
			longPollRequest.ProbeModifyRequest += lineSeparator
		}
		longPollRequest.ProbeModifyRequest += strings.Join([]string{infoParam.DataID, infoParam.Group, infoParam.ContentMD5, d.option.tenant}, wordSeparator)
	}
	headerSetters := []headerSetter{
		d.withLongPollingTimeout(),
		d.withSignature(d.option.tenant, ""),
	}
	header := make(http.Header)
	for _, setter := range headerSetters {
		if err := setter(header); err != nil {
			return nil, err
		}
	}

	request := d.c.NewRequest().
		WithPath(acmLongPull.String(ip)).
		WithFormURLEncodedBody(longPollRequest).
		WithHeader(header).
		Post()
	response, err := d.c.Do(context.TODO(), request)
	if err != nil {
		return nil, err
	}
	if !response.Success() {
		return nil, errors.New(response.String())
	}
	var ret []info.Info

	s, err := url.QueryUnescape(response.String())
	if err != nil {
		return nil, errors.New("unescape fail")
	}

	ss := strings.Split(s, lineSeparator)
	for _, s := range ss {
		tt := strings.Split(s, wordSeparator)
		if len(tt) != 3 {
			continue
		}
		i := info.Info{
			DataID: tt[0],
			Group:  tt[1],
		}
		ret = append(ret, i)
	}
	return ret, nil
}
