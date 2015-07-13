package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Add struct {
	*command
}

func NewAdd() *Add {
	return &Add{newCommand()}
}

func (c *Add) Name() string {
	return "add"
}

func (c *Add) Description() string {
	return fmt.Sprintf("Asks the deamon for general information and the specifics of the current timer, it allows for arbritary formatting of the current time measurement.")
}

func (c *Add) Usage() string {
	return "Show info on the timer for this repository"
}

func (c *Add) Flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{Name: "all,A", Usage: "Stage time spend on all files currently staged in git"},
	}
}

func (c *Add) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Add) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	staged, err := vc.Staging()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to get staged files from the VCS: {{err}}"), err)
	}

	client := NewClient()

	if ctx.Bool("all") {
		c.Printf("Staging time for %d files:", len(staged))
		for _, f := range staged {
			rel, err := filepath.Rel(vc.Root(), f.Path())
			if err != nil {
				c.Printf("Failed to determine relative path for '%s'", f.Path())
			}

			c.Printf(" %s: %s", f.Date(), rel)
		}

		err := client.StageTimer(vc.Root(), staged)
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to stage files: {{err}}"), err)
		}
	} else {
		c.Fatal("Not yet implemented")
	}

	return nil
}
