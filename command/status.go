package command

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/config"
	"github.com/timeglass/glass/vcs"
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
	return fmt.Sprintf("Asks the deamon for general information and the specifics of the current timer, it allows for arbritary formatting of the current time measurement.")
}

func (c *Status) Usage() string {
	return "Show info on the timer for this repository"
}

func (c *Status) Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "template,t", Value: "", Usage: "a template that allows for arbritary formatting of the time output"},
		cli.BoolFlag{Name: "commit-template", Usage: "use the commit template from the configuration, this overwrites and custom template using -t"},
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

	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	conf, err := config.ReadConfig(vc.Root())
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to read configuration: {{err}}"), err)
	}

	client := NewClient()

	//fetch information on overall daemon
	c.Printf("Fetching daemon info...")
	dinfo, err := client.Info()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to fetch daemon info, is it installed and running?: {{err}}"), err)
	}

	curr, _ := strconv.Atoi(strings.Replace(dinfo["version"].(string), ".", "", 2))
	recent, _ := strconv.Atoi(strings.Replace(dinfo["newest_version"].(string), ".", "", 2))
	if curr != 0 && recent > curr {
		c.Println("A new version is available, please upgrade: https://github.com/timeglass/glass/releases")
	}

	//fetch information on the timer specific to this directory
	c.Printf("Fetching timer info...")
	timer, err := client.ReadTimer(vc.Root())
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to fetch timer: {{err}}"), err)
	}

	if reason := timer.HasFailed(); reason != "" {
		c.Printf("Timer has failed: %s", reason)
	} else {
		if timer.IsPaused() {
			c.Printf("Timer is currently: PAUSED")
		} else {
			c.Printf("Timer is currently: RUNNING")
		}
	}

	tmpls := ctx.String("template")
	if ctx.Bool("commit-template") {
		tmpls = conf.CommitMessage
	}

	//we got some template specified
	if tmpls != "" {

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

	} else {
		//just print
		c.Printf("Timer reads: %s", timer.Time())
	}

	return nil
}
