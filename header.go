package aliacm

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	longPullingTimeoutHeader = "longPullingTimeout"
	spssAccessKeyHeader      = "Spas-AccessKey"
	timeStampHeader          = "timeStamp"
	spasSignatureHeader      = "Spas-Signature"
)

type headerSetter func(header http.Header) error

func (d *Diamond) withLongPollingTimeout() headerSetter {
	return func(header http.Header) error {
		if header == nil {
			header = make(http.Header)
		}
		header.Set(longPullingTimeoutHeader, "30000")
		return nil
	}
}

func (d *Diamond) withUsual(tenant, group string) headerSetter {
	return func(header http.Header) error {
		if header == nil {
			header = make(http.Header)
		}
		now := timeInMilli()
		var toSignList []string
		toSignList = append(toSignList, tenant, group, strconv.FormatInt(now, 10))
		str := strings.Join(toSignList, "+")
		signature, err := HMACSHA1Encrypt(str, d.option.secretKey)
		if err != nil {
			return err
		}
		header.Set(spssAccessKeyHeader, d.option.accessKey)
		header.Set(timeStampHeader, strconv.FormatInt(now, 10))
		header.Set(spasSignatureHeader, signature)
		return nil
	}
}
