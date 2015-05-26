package command

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

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
	return fmt.Sprintf("Reads time data for a list of commits provided to the command over STDIN. It expects one commit per line and the can be specified in any format that the underlying VCS accepts (refs, hashes, short hashes etc)")
}

func (c *Sum) Usage() string {
	return "Collect time measurements from a series of commits"
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

	//retrieve commits through piped stdin or arguments
	commits := []string{}
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			commits = append(commits, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
	} else {
		if ctx.Args().First() != "" {
			commits = append(commits, ctx.Args().First())
		}
		commits = append(commits, ctx.Args().Tail()...)
	}

	if len(commits) == 0 {
		return fmt.Errorf("Please provide at least one commit through STDIN or as an argument.")
	}

	//map time data from the vcs
	list := []vcs.TimeData{}
	for _, c := range commits {
		data, err := vc.Show(c)
		if err != nil {
			if err == vcs.ErrNoCommitTimeData {
				//ignore if a commit has no time attached
				continue
			}

			return errwrap.Wrapf(fmt.Sprintf("Failed to show time notes for '%s': {{err}}", c), err)
		}

		list = append(list, data)
	}

	//reduce to output
	var total time.Duration
	for _, data := range list {
		total += data.Total()
	}

	fmt.Println(total)
	return nil
}
