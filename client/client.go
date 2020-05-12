package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	url        string
	username   string
	password   string
	insecure   bool
	httpClient *http.Client
}

type ClientResponse struct {
	StatusCode int
	Headers    http.Header
	Body       string
}

// NewClient creates common settings
func NewClient(url string, username string, password string, insecure bool) *Client {

	return &Client{
		url:        url,
		username:   username,
		password:   password,
		insecure:   insecure,
		httpClient: &http.Client{},
	}
}

func (c *Client) SendRequestFull(method string, path string, payload interface{}, statusCode int) (value *ClientResponse, err error) {
	url := c.url + path
	client := &http.Client{}

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(payload)
	if err != nil {
		return nil, err
	}

	if c.insecure == true {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}

	req, err := http.NewRequest(method, url, b)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	strbody := string(body)

	if statusCode != 0 {
		if resp.StatusCode != statusCode {
			return nil, fmt.Errorf("[ERROR] unexpected status code got: %v expected: %v \n %v", resp.StatusCode, statusCode, strbody)
		}
	}

	return &ClientResponse{
		StatusCode: resp.StatusCode,
		Headers: resp.Header,
		Body: strbody,
	}, nil
}

func (c *Client) SendRequest(method string, path string, payload interface{}, statusCode int) (value string, err error) {
	resp, err := c.SendRequestFull(method, path, payload, statusCode)
	if err != nil {
		return "", err
	}
	return resp.Body, err
}