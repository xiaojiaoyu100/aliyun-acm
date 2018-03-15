package acm

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab.xinghuolive.com/golang/utils/error"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// GetConfig gets config value in UTF-8 from ACM.
func (c Client) GetConfig(group string, dataID string) string {
	resp := c.get(group, dataID)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder()))
	e.Panic(err)

	return string(body)
}

func (c Client) getConfigWithMD5(group string, dataID string) (string, string) {
	resp := c.get(group, dataID)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	e.Panic(err)

	contentMD5 := fmt.Sprintf("%x", md5.Sum(body))
	decodedBody, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(body), simplifiedchinese.GBK.NewDecoder()))
	e.Panic(err)

	return string(decodedBody), contentMD5
}

func (c Client) get(group string, dataID string) *http.Response {
	url := fmt.Sprintf("http://%s/diamond-server/config.co?dataId=%s&group=%s&tenant=%s",
		c.ServerIP, dataID, group, c.Tenant)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	e.Panic(err)

	ts := strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond)))
	signText := strings.Join([]string{c.Tenant, group, ts}, "+")

	sign := getSign(signText, c.SecretKey)

	req.Header.Set(headerAccessKey, c.AccessKey)
	req.Header.Set(headerTS, ts)
	req.Header.Set(headerSignature, sign)

	resp, err := httpClient.Do(req)
	e.Panic(err)

	return resp
}
