package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Stage struct {
	*command
}

func NewStage() *Stage {
	return &Stage{newCommand()}
}

func (c *Stage) Name() string {
	return "stage"
}

func (c *Stage) Description() string {
	return fmt.Sprintf("...")
}

func (c *Stage) Usage() string {
	return "..."
}

func (c *Stage) Flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{Name: "all,A", Usage: "Stage time spend on all files currently staged in git"},
	}
}

func (c *Stage) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Stage) Run(ctx *cli.Context) error {
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
