package command

import (
	"bytes"
	// "encoding/json"
	// "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	// "strings"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type Client struct {
	endpoint string
	*http.Client
}

func NewClient() *Client {
	return &Client{
		endpoint: "http://127.0.0.1:3838",
		Client: &http.Client{
			Timeout: time.Duration(400 * time.Millisecond),
		},
	}
}

func (c *Client) Call(method string, params url.Values) ([]byte, error) {
	loc := fmt.Sprintf("%s/api/%s?%s", c.endpoint, method, params.Encode())
	resp, err := c.Get(loc)
	if err != nil {
		return []byte{}, errwrap.Wrapf(fmt.Sprintf("Failed to GET %s: {{err}}", loc), err)
	}

	body := bytes.NewBuffer(nil)
	defer resp.Body.Close()
	_, err = io.Copy(body, resp.Body)
	if err != nil {
		return body.Bytes(), errwrap.Wrapf(fmt.Sprintf("Failed to buffer response body: {{err}}"), err)
	}

	if resp.StatusCode > 299 {
		return body.Bytes(), fmt.Errorf("Unexpected StatusCode returned from Deamon: '%d', body: '%s'", resp.StatusCode, body.String())
	}

	return body.Bytes(), nil
}

func (c *Client) CreateTimer(dir string) error {
	params := url.Values{}
	params.Set("dir", dir)

	_, err := c.Call("timers.create", params)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed call http endpoint 'timers.create' with '%s': {{err}}", dir), err)
	}

	return nil
}
