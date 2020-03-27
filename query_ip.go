package aliacm

import (
	"context"
	"errors"
	"strings"
)

// QueryIP 查询配置
func (d *Diamond) QueryIP() (string, error) {
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
		return ips[idx], nil
	}
	return "", nil
}
