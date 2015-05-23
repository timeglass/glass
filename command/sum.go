package command

import (
	"bufio"
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Sum struct {
	*command
}

func NewSum() *Sum {
	return &Sum{newCommand()}
}

func (c *Sum) Name() string {
	return "sum"
}

func (c *Sum) Description() string {
	return fmt.Sprintf("<description>")
}

func (c *Sum) Usage() string {
	return "<usage>"
}

func (c *Sum) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Sum) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Sum) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	//write the vcs
	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	_ = vc

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			fmt.Println("line:", line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}
