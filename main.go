package main

import (
	"fmt"
	"os"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"

	"github.com/timeglass/glass/command"
)

type Command interface {
	Name() string
	Description() string
	Usage() string
	Run(c *cli.Context) error
	Action() func(ctx *cli.Context)
	Flags() []cli.Flag
}

var Version = "0.0.0"
var Build = "gobuild"

func main() {
	app := cli.NewApp()
	app.Author = "Ad van der Veer"
	app.Email = "advanderveer@gmail.com"
	app.Name = "Timeglass"
	app.Usage = "Automated time tracking for code repositories"
	app.Version = fmt.Sprintf("%s (%s)", Version, Build)

	cmds := []Command{
		command.NewInstall(),   //install daemon and start service
		command.NewUninstall(), //stop daemon and uninstall service
		command.NewInit(),      //write hooks, create timer and pull time data
		command.NewStart(),     //create timer for current directory, start measuring
		command.NewPause(),     //pause timer for the current directory, restart on file activity
		command.NewStatus(),    //fetch info of the timer for the current directory
		command.NewReset(),     //reset the timer to 0s
		command.NewStop(),      //remove timer for current directory, discarding meaurement
		command.NewPush(),      //push notes branch to remote
		command.NewPull(),      //pull notes branch from remote
		command.NewPunch(),     //persist time measurement to current HEAD commit
		command.NewSum(),       //sum total time of each commit given
	}

	for _, c := range cmds {
		app.Commands = append(app.Commands, cli.Command{
			Name:        c.Name(),
			Usage:       c.Usage(),
			Action:      c.Action(),
			Description: c.Description(),
			Flags:       c.Flags(),
		})
	}

	app.Run(os.Args)
}
