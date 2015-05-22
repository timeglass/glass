package command

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

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
	return ""
}

func (c *Status) Usage() string {
	return "Show info on the running timer"
}

func (c *Status) Flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "time-only",
			Usage: "Only display the time",
		}}
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
	info, err := m.ReadDaemonInfo()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to get Daemon address: {{err}}"), err)
	}

	conf, err := m.ReadConfig()
	if err != nil {
		//@todo find a more elegant way to 'print' this for script usage
		if ctx.Bool("time-only") {
			return nil
		}

		return errwrap.Wrapf(fmt.Sprintf("Failed to read configuration: {{err}}"), err)
	}

	client := NewClient(info)
	status, err := client.GetStatus()
	if err != nil {
		if err == ErrDaemonDown {
			//if called from hook, don't interrupt
			if ctx.Bool("time-only") {
				return nil
			}

			return errwrap.Wrapf(fmt.Sprintf("No timer appears to be running for '%s': {{err}}", dir), err)
		} else {
			return err
		}
	}

	t, err := time.ParseDuration(status.Time)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to parse '%s' as a time duration: {{err}}", status.Time), err)
	}

	if !ctx.Bool("time-only") {
		//simple semver check
		curr, _ := strconv.Atoi(strings.Replace(status.CurrentVersion, ".", "", 2))
		recent, _ := strconv.Atoi(strings.Replace(status.MostRecentVersion, ".", "", 2))
		if curr != 0 && recent > curr {
			fmt.Println("A new version of Timeglass is available, please upgrade from https://github.com/timeglass/glass/releases.")
		}
	} else if t.Seconds() == 0 {
		//for script usage we return nothing when there has zero
		//time elapsed, this prevents empty bracke
		return nil
	}

	//parse temlate and only report error if we're talking to a human
	tmpl, err := template.New("commit-msg").Parse(conf.CommitMessage)
	if err != nil {
		//@todo find a more elegant way to 'print' this for script usage
		if ctx.Bool("time-only") {
			return nil
		} else {
			return errwrap.Wrapf(fmt.Sprintf("Failed to parse commit_message: '%s' in configuration as a text/template: {{err}}", conf.CommitMessage), err)
		}
	}

	//execute template and write to stdout
	err = tmpl.Execute(os.Stdout, t)
	if err != nil {
		//@todo find a more elegant way to 'print' this for script usage
		if ctx.Bool("time-only") {
			return nil
		} else {
			return errwrap.Wrapf(fmt.Sprintf("Failed to execute commit_message: template for time '%s': {{err}}", t), err)
		}
	}

	//end with newline if we're printing for a human
	if !ctx.Bool("time-only") {
		fmt.Println()
	}

	return nil
}
