package command

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"
	"github.com/mattn/go-isatty"

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

	var total time.Duration
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()

			data, err := vc.Show(line)
			if err != nil {
				if err == vcs.ErrNoCommitTimeData {
					//ignore if a commit has no time attached
					continue
				}

				return errwrap.Wrapf(fmt.Sprintf("Failed to show time notes for '%s': {{err}}", line), err)
			}

			total += data.Total()
		}
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println(total)
	return nil
}
