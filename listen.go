package acm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab.xinghuolive.com/golang/utils/error"
)

// ConfigHandler is a shortcut of handler function.
type ConfigHandler func(newValue string)

// Listen calls config handler when config updates.
func (c Client) Listen(group string, dataID string, handler ConfigHandler) {
	lastMD5 := ""
	for {
		if c.isUpdated(group, dataID, lastMD5) {
			var newValue string
			newValue, lastMD5 = c.getConfigWithMD5(group, dataID)
			handler(newValue)
		}
	}
}

func (c Client) isUpdated(group string, dataID string, lastMD5 string) bool {
	url := fmt.Sprintf("http://%s/diamond-server/config.co", c.ServerIP)
	content := strings.Join([]string{dataID, group, lastMD5, c.Tenant}, string(rune(2))) + string(rune(1))
	params := fmt.Sprintf("Probe-Modify-Request=%s", content)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(params))
	e.Panic(err)

	req.Header.Set(headerTimeout, "30000")
	req.Header.Set(headerAccessKey, c.AccessKey)
	req.Header.Set(headerTS, strconv.Itoa(int(time.Now().UnixNano()/int64(time.Millisecond))))
	req.Header.Set(headerSignature, getSign(content, c.SecretKey))

	resp, err := httpClient.Do(req)
	e.Panic(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	e.Panic(err)

	return len(body) != 0
}
