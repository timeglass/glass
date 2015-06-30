package command

import (
	"fmt"
	"os"
	"text/template"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/model"
)

type Status struct {
	*command
}

func NewStatus() *Status {
	return &Status{newCommand()}
}

func (c *Status) Name() string {
	return "status"
}

func (c *Status) Description() string {
	return fmt.Sprintf("...")
}

func (c *Status) Usage() string {
	return "..."
}

func (c *Status) Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "template,t", Value: "", Usage: "a template that allows for arbritary formatting of the time output"},
	}
}

func (c *Status) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Status) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	m := model.New(dir)
	conf, err := m.ReadConfig()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to read configuration: {{err}}"), err)
	}

	c.Printf("Fetching timer info...")

	client := NewClient()
	timer, err := client.ReadTimer(dir)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to fetch timer: {{err}}"), err)
	}

	tmpls := ctx.String("template")
	if tmpls == "" {
		tmpls = conf.CommitMessage
	}

	//parse temlate and only report error if we're talking to a human
	tmpl, err := template.New("commit-msg").Parse(tmpls)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to parse commit_message: '%s' in configuration as a text/template: {{err}}", conf.CommitMessage), err)
	}

	//execute template and write to stdout
	err = tmpl.Execute(os.Stdout, timer.Time())
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to execute commit_message: template for time '%s': {{err}}", timer.Time()), err)
	}

	return nil
}
