package rest

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"bitbucket.org/rbergman/go-hipchat-connect/store"
)

type MediaType string

const JSONType MediaType = "application/json"

type Client struct {
	BaseURL string
	Tokens  store.Store
}

func NewClient(baseURL string, s store.Store) *Client {
	if s == nil {
		s = store.NewDefaultMemoryStore()
	}
	return &Client{
		BaseURL: baseURL,
		Tokens:  s,
	}
}

func (c *Client) Post(url string, headers map[string]string, r io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, r)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		contents, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(fmt.Sprintf("Unexpected response: %d \n %s", resp.StatusCode, contents))
	}

	return resp, nil
}

func checkJSONResponse(res *http.Response, expect int) error {
	if res.StatusCode != expect {
		defer res.Body.Close()
		contents, _ := ioutil.ReadAll(res.Body)
		return errors.New(fmt.Sprintf("Unexpected response status: %d \n %s", res.StatusCode, contents))
	}

	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		return errors.New("Unexpected content-type: " + contentType)
	}

	if res.Body == nil {
		return errors.New("Expected JSON response body")
	}

	return nil
}
