package command

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/advanderveer/timer/model"
)

type Start struct {
	*command
}

func NewStart() *Start {
	return &Start{newCommand()}
}

func (c *Start) Name() string {
	return "start"
}

func (c *Start) Description() string {
	return fmt.Sprintf("<description>")
}

func (c *Start) Usage() string {
	return "<usage>"
}

func (c *Start) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Start) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Start) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	m := model.New(dir)
	d, err := m.ReadDaemonInfo()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to get Daemon address: {{err}}"), err)
	}

	to := time.Duration(100 * time.Millisecond)
	client := http.Client{
		Timeout: to,
	}

	resp, err := client.Get(fmt.Sprintf("http://%s/timer.start", d.Addr))
	if err != nil {
		//@todo start hook script, it doesn't appear to be running
		log.Fatal(err)
	} else if resp.StatusCode != 200 {
		log.Fatalf("Unexpected StatusCode from Deamon: '%d'", resp.StatusCode)
	}

	return nil
}
