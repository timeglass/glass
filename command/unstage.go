package command

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Unstage struct {
	*command
}

func NewUnstage() *Unstage {
	return &Unstage{newCommand()}
}

func (c *Unstage) Name() string {
	return "unstage"
}

func (c *Unstage) Description() string {
	return fmt.Sprintf("...")
}

func (c *Unstage) Usage() string {
	return "..."
}

func (c *Unstage) Flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{Name: "all,A", Usage: "Unstage for all files"},
	}
}

func (c *Unstage) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Unstage) Run(ctx *cli.Context) error {
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

		c.Printf("Unstaging time for %d files:", len(staged))
		for _, f := range staged {
			rel, err := filepath.Rel(vc.Root(), f.Path())
			if err != nil {
				c.Printf("Failed to determine relative path for '%s'", f.Path())
			}

			c.Printf("- %s", rel)

			//we set the staging date in the past (unix epoch) to effectively unstage
			f.SetDate(time.Unix(0, 0))
		}

		err := client.StageTimer(vc.Root(), staged)
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to unstage files: {{err}}"), err)
		}

	} else {
		c.Fatal("Not yet implemented")
	}

	return nil
}
