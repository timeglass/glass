package command

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
)

type command struct {
	*log.Logger
}

func newCommand() *command {
	return &command{
		log.New(os.Stderr, "glass: ", log.Ltime),
	}
}

func (c *command) Action(fn func(c *cli.Context) error) func(ctx *cli.Context) {
	return func(ctx *cli.Context) {
		if ctx.GlobalBool("silent") {
			c.Logger = log.New(ioutil.Discard, "", 0)
		}

		err := fn(ctx)
		if err != nil {
			c.Fatal(err)
			return
		}
	}
}
