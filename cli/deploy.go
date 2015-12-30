package cli

import (
	"github.com/ch3lo/overlord/api"
	"github.com/codegangsta/cli"
)

func deployFlags() []cli.Flag {
	return []cli.Flag{}
}

func deployBefore(c *cli.Context) error {
	return nil
}

func deployCmd(c *cli.Context) {
	api.Server()
}
