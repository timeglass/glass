package command

import (
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
		err := fn(ctx)
		if err != nil {
			c.Fatalf("Error: %s", err)
			return
		}
	}
}
