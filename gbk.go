package aliacm

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// GbkToUtf8 converts gbk encoded bytes to utf-8 encoded bytes.
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	ret, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
