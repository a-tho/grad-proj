package main

import (
	"path"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/a-tho/grad-proj/internal/generator"
)

func main() {
	app := cli.NewApp()

	app.Commands = []*cli.Command{
		{
			Name:      "server",
			Usage:     "generate server transport for given interface",
			UsageText: "xua server",
			Action:    actionServer,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "in",
					Usage: "path to package with input interface",
				},
				&cli.StringFlag{
					Name:  "out",
					Usage: "path for output transport",
				},
			},
		},
		// {
		// 	Name:      "client",
		// 	Usage:     "generate client for given interface",
		// 	UsageText: "xua client",
		// 	Action:    actionServer,
		// 	Flags: []cli.Flag{
		// 		&cli.StringFlag{
		// 			Name:  "in",
		// 			Usage: "path to package with input interface",
		// 		},
		// 		&cli.StringFlag{
		// 			Name:  "out",
		// 			Usage: "path for command output",
		// 		},
		// 	},
		// },
	}
}

func actionServer(c *cli.Context) (err error) {
	t, err := generator.NewTransport(log.Logger, c.String("in"))
	if err != nil {
		return
	}

	out, _ := path.Split(c.String("in"))
	out = path.Join(out, "transport")
	if c.String("out") != "" {
		out = c.String("out")
	}

	err = t.GenerateServer(out)
	if err != nil {
		return
	}

	return
}

// func actionClient(c *cli.Context) (err error) {}
