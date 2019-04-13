package acm

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// GetConfig gets config value in UTF-8 from ACM.
func (c Client) GetConfig(group string, dataID string) (string, error) {
	resp, err := c.get(group, dataID)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder()))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}

	return string(body), nil
}

func (c Client) getConfigWithMD5(group string, dataID string) (string, string, error) {
	resp, err := c.get(group, dataID)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	contentMD5 := fmt.Sprintf("%x", md5.Sum(body))
	decodedBody, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(body), simplifiedchinese.GBK.NewDecoder()))
	if err != nil {
		return "", "", err
	}

	return string(decodedBody), contentMD5, nil
}

func (c Client) get(group string, dataID string) (*http.Response, error) {
	url := fmt.Sprintf("http://%s/diamond-server/config.co?dataId=%s&group=%s&tenant=%s",
		c.ServerIP, dataID, group, c.Tenant)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	ts := strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond)))
	signText := strings.Join([]string{c.Tenant, group, ts}, "+")

	sign := getSign(signText, c.SecretKey)

	req.Header.Set(headerAccessKey, c.AccessKey)
	req.Header.Set(headerTS, ts)
	req.Header.Set(headerSignature, sign)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
