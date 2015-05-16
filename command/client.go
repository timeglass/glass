package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/model"
)

var ErrDaemonDown = errors.New("Daemon doesn't appears to be running.")

type Client struct {
	info *model.Daemon

	*http.Client
}

func NewClient(info *model.Daemon) *Client {
	return &Client{
		info: info,
		Client: &http.Client{
			Timeout: time.Duration(100 * time.Millisecond),
		},
	}
}

func (c *Client) Call(method string) error {
	resp, err := c.Get(fmt.Sprintf("http://%s/%s", c.info.Addr, method))
	if err != nil {
		return ErrDaemonDown
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("Unexpected StatusCode from Daemon: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Lap() (time.Duration, error) {
	resp, err := c.Get(fmt.Sprintf("http://%s/timer.lap", c.info.Addr))
	if err != nil {
		return 0, ErrDaemonDown
	} else if resp.StatusCode != 200 {
		return 0, fmt.Errorf("Unexpected StatusCode from Daemon: %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	status := struct {
		Time string
	}{}

	err = dec.Decode(&status)
	if err != nil {
		return 0, errwrap.Wrapf("Failed to decode json response: {{err}}", err)
	}

	d, err := time.ParseDuration(status.Time)
	if err != nil {
		return 0, errwrap.Wrapf(fmt.Sprintf("Failed to parse '%s' as a time duration: {{err}}", status.Time), err)
	}

	return d, nil
}

func (c *Client) Split() (time.Duration, error) {
	resp, err := c.Get(fmt.Sprintf("http://%s/timer.status", c.info.Addr))
	if err != nil {
		return 0, ErrDaemonDown
	} else if resp.StatusCode != 200 {
		return 0, fmt.Errorf("Unexpected StatusCode from Daemon: %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	status := struct {
		Time string
	}{}

	err = dec.Decode(&status)
	if err != nil {
		return 0, errwrap.Wrapf("Failed to decode json response: {{err}}", err)
	}

	d, err := time.ParseDuration(status.Time)
	if err != nil {
		return 0, errwrap.Wrapf(fmt.Sprintf("Failed to parse '%s' as a time duration: {{err}}", status.Time), err)
	}

	return d, nil
}
