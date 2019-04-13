package acm

import (
	"io/ioutil"
	"strings"
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
func GetClient(endpoint string, tenant string, accessKey string, secretKey string) (*Client, error) {
	client := Client{
		EndPoint:  endpoint,
		Tenant:    tenant,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	resp, err := httpClient.Get("http://" + endpoint + "/diamond-server/diamond")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	client.ServerIP = strings.Split(string(body), "\n")[0] + ":8080"

	return &client, nil
}
