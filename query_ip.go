package aliacm

import (
	"context"
	"errors"
	"strings"
	"time"
)

const ip = "ip"

// QueryIP 查询配置
func (d *Diamond) QueryIP() (string, error) {
	itf, _ := d.cache.Get(ip)
	ip, ok := itf.(string)
	if ok && len(ip) > 0 {
		return ip, nil
	}

	request := d.c.NewRequest().
		WithTimeout(apiTimeout).
		WithPath(acmIP.String(d.option.addr)).
		Get()
	response, err := d.c.Do(context.TODO(), request)
	if err != nil {
		return "", err
	}
	if !response.Success() {
		return "", errors.New(response.String())
	}
	ips := strings.Split(strings.TrimSpace(response.String()), "\n")
	if len(ips) > 0 {
		idx := d.r.Intn(len(ips))
		_ = d.cache.Set(ip, ips[idx], time.Second*60)
		return ips[idx], nil
	}
	return "", nil
}
