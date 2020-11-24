package main

import "github.com/urfave/cli/v2"

func scanRadiusSvc(ctx *cli.Context) error {
	handler := scanResultHandler{}
	return startConsumer(handler)
}
