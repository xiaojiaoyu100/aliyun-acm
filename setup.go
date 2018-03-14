package acm

import (
	"fmt"
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

var client Client

// Setup sets configs of ACM.
func Setup(endpoint string, tenant string, accessKey string, secretKey string) {
	client = Client{
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
	fmt.Println(client)
}
