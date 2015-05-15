package command

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/advanderveer/timer/model"
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
