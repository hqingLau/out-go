package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	var language string

	app := cli.NewApp()

	app.Usage = "hahaha"
	app.Authors = []*cli.Author{{
		Name:  "hqinglau",
		Email: "hqinglau@gmail.com",
	}}

	app.Action = func(ctx *cli.Context) error {
		fmt.Println("Boom!")
		fmt.Println(language)
		return nil
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "lang",
			Value:       "english",
			Usage:       "language for the greeting",
			Destination: &language,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
