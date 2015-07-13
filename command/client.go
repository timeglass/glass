package command

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/timer"
	"github.com/timeglass/glass/vcs"
)

var ErrRequestFailed = errors.New("Couldn't reach background service, did you install it using 'glass install'?")
var ErrTimerNotFound = errors.New("Couldn't find timer for this project, did you start one using 'glass init' or 'glass start'?")

type Client struct {
	endpoint string
	*http.Client
}

func NewClient() *Client {
	return &Client{
		endpoint: "http://127.0.0.1:3838",
		Client:   &http.Client{},
	}
}

func (c *Client) Call(method string, params url.Values) ([]byte, error) {
	loc := fmt.Sprintf("%s/api/%s", c.endpoint, method)
	resp, err := c.Post(loc, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(params.Encode())))
	if err != nil {
		return nil, ErrRequestFailed
	}

	body := bytes.NewBuffer(nil)
	defer resp.Body.Close()
	_, err = io.Copy(body, resp.Body)
	if err != nil {
		return body.Bytes(), errwrap.Wrapf(fmt.Sprintf("Failed to buffer response body: {{err}}"), err)
	}

	if resp.StatusCode > 299 {
		errresp := &struct {
			Error string
		}{}

		err := json.Unmarshal(body.Bytes(), &errresp)
		if err != nil || errresp.Error == "" {
			return body.Bytes(), fmt.Errorf("Unexpected StatusCode returned from Deamon: '%d', body: '%s'", resp.StatusCode, body.String())
		} else if strings.Contains(errresp.Error, "No known timer") {
			return body.Bytes(), ErrTimerNotFound
		}

		return body.Bytes(), fmt.Errorf(errresp.Error)
	}

	return body.Bytes(), nil
}

func (c *Client) Info() (map[string]interface{}, error) {
	data, err := c.Call("", url.Values{})
	if err != nil {
		return nil, err
	}

	v := map[string]interface{}{}
	err = json.Unmarshal(data, &v)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to deserialize '%s' into map: {{err}}", data), err)
	}

	return v, nil
}

func (c *Client) CreateTimer(dir string) error {
	params := url.Values{}
	params.Set("dir", dir)

	_, err := c.Call("timers.create", params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteTimer(dir string) error {
	params := url.Values{}
	params.Set("dir", dir)

	_, err := c.Call("timers.delete", params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ResetTimer(dir string) error {
	params := url.Values{}
	params.Set("dir", dir)

	_, err := c.Call("timers.reset", params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) StageTimer(dir string, files map[string]*vcs.StagedFile) error {
	params := url.Values{}
	params.Set("dir", dir)

	for _, sf := range files {
		params.Add("files", fmt.Sprintf("%s:%d", sf.Path(), sf.Date().UnixNano()))
	}

	data, err := c.Call("timers.stage", params)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}

func (c *Client) PauseTimer(dir string) error {
	params := url.Values{}
	params.Set("dir", dir)

	_, err := c.Call("timers.pause", params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ReadTimer(dir string) (*timer.Timer, error) {
	timers := []*timer.Timer{}
	params := url.Values{}
	params.Set("dir", dir)

	data, err := c.Call("timers.info", params)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &timers)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to deserialize '%s' into a list of timers: {{err}}", data), err)
	}

	if len(timers) < 1 {
		return nil, fmt.Errorf("Expected at least one timer from the daemon")
	}

	return timers[0], nil
}
