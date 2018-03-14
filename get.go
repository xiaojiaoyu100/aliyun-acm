package acm

import (
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

// Get gets config value from ACM.
func Get(group string, dataID string) string {
	url := fmt.Sprintf("http://%s/diamond-server/config.co?dataId=%s&group=%s&tenant=%s",
		client.ServerIP, dataID, group, client.Tenant)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	e.Panic(err)

	ts := strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond)))
	signText := strings.Join([]string{client.Tenant, group, ts}, "+")

	sign := getSign(signText, client.SecretKey)

	req.Header.Set(headerAccessKey, client.AccessKey)
	req.Header.Set(headerTS, ts)
	req.Header.Set(headerSignature, sign)

	resp, err := httpClient.Do(req)
	e.Panic(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder()))
	e.Panic(err)

	return string(body)
}
