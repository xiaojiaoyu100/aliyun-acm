package aliacm

import (
	"errors"
	"strings"
)

// QueryIP 查询配置
func (d *Diamond) QueryIP() (string, error) {
	request := d.c.NewRequest().
		WithTimeout(apiTimeout).
		WithPath(acmIP.String(d.option.addr)).
		Get()
	response, err := d.c.Do(request)
	if err != nil {
		return "", err
	}
	if !response.Success() {
		return "", errors.New(response.String())
	}
	ips := strings.Split(response.String(), "\n")
	// TODO: randomly select one from ip list?
	return ips[0], nil
}
