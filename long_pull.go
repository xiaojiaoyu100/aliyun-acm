package aliacm

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const (
	wordSeparator = string(rune(2))
	lineSeparator = string(rune(1))
)

// LongPull 监听配置
func (d *Diamond) LongPull(unit Unit, contentMD5 string) (string, error) {
	ip, err := d.QueryIP()
	if err != nil {
		return "", err
	}
	switch contentMD5 {
	case "":
		args := new(GetConfigRequest)
		args.Tenant = d.option.tenant
		args.Group = unit.Group
		args.DataID = unit.DataID
		content, err := d.GetConfig(args)
		if err != nil {
			return "", err
		}
		contentMD5 = Md5(string(content))
		config := Config{
			Content: content,
		}
		unit.ch <- config
		return contentMD5, nil
	default:
		headerSetters := []headerSetter{
			d.withLongPollingTimeout(),
			d.withUsual(d.option.tenant, unit.Group),
		}
		header := make(http.Header)
		for _, setter := range headerSetters {
			if err := setter(header); err != nil {
				return "", err
			}
		}
		var longPollRequest struct {
			ProbeModifyRequest string `url:"Probe-Modify-Request"`
		}
		longPollRequest.ProbeModifyRequest = strings.Join([]string{unit.DataID, unit.Group, contentMD5, d.option.tenant}, wordSeparator) + lineSeparator
		request := d.c.NewRequest().
			WithPath(acmLongPull.String(ip)).
			WithFormURLEncodedBody(longPollRequest).
			WithHeader(header).
			Post()
		response, err := d.c.Do(request)
		if err != nil {
			return "", err
		}
		switch response.StatusCode() {
		case http.StatusServiceUnavailable:
			return "", serviceUnavailableErr
		case http.StatusInternalServerError:
			return "", internalServerErr
		}
		if !response.Success() {
			return "", errors.New(response.String())
		}
		ret := url.QueryEscape(strings.Join([]string{unit.DataID, unit.Group, d.option.tenant}, wordSeparator) + lineSeparator)
		if ret == strings.TrimSpace(response.String()) {
			args := new(GetConfigRequest)
			args.Tenant = d.option.tenant
			args.Group = unit.Group
			args.DataID = unit.DataID
			content, err := d.GetConfig(args)
			if err != nil {
				return "", err
			}
			contentMD5 := Md5(string(content))
			config := Config{
				Content: content,
			}
			unit.ch <- config
			return contentMD5, nil
		}
		return contentMD5, nil
	}
}
