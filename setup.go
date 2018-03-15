package acm

import (
	"io/ioutil"
	"strings"

	"gitlab.xinghuolive.com/golang/utils/error"
)

// Client contains ACM configs.
type Client struct {
	EndPoint  string
	Tenant    string
	AccessKey string
	SecretKey string
	ServerIP  string
}

// GetClient sets configs of ACM and return client struct.
func GetClient(endpoint string, tenant string, accessKey string, secretKey string) Client {
	client := Client{
		EndPoint:  endpoint,
		Tenant:    tenant,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	resp, err := httpClient.Get("http://" + endpoint + "/diamond-server/diamond")
	e.Panic(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	e.Panic(err)

	client.ServerIP = strings.TrimSpace(string(body)) + ":8080"

	return client
}
