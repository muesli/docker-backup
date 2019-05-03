/*
 * A tool to create & restore full backups of Docker containers
 *     Copyright (c) 2019, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var (
	cli *client.Client
	ctx = context.Background()

	RootCmd = &cobra.Command{
		Use:           "docker-backup",
		Short:         "docker-backup creates or restores backups of Docker containers",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func main() {
	var err error
	// cli, err = client.NewEnvClient()
	// cli, err = client.NewClientWithOpts(client.FromEnv)
	cli, err = client.NewClientWithOpts(client.WithVersion("1.36"))
	if err != nil {
		panic(err)
	}

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
