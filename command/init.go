package command

import (
	"fmt"
	"os"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Init struct {
	*command
}

func NewInit() *Init {
	return &Init{newCommand()}
}

func (c *Init) Name() string {
	return "init"
}

func (c *Init) Description() string {
	return fmt.Sprintf("Install hooks for the current repository, if hooks already exists they are truncated and rewritten.")
}

func (c *Init) Usage() string {
	return "Install Timeglass for the current repository"
}

func (c *Init) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Init) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Init) Run(ctx *cli.Context) error {
	c.Println("Writing version control hooks...")
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	err = vc.Hook()
	if err != nil {
		return errwrap.Wrapf("Failed to write hooks: {{err}}", err)
	}

	c.Println("Hooks written!")
	err = NewStart().Run(ctx)
	if err != nil {
		return err
	}

	err = NewPull().Run(ctx)
	if err != nil {
		if errwrap.Contains(err, vcs.ErrNoRemote.Error()) {
			c.Println("No remote found, skipping pull")
		} else {
			return err
		}
	}

	return nil
}
