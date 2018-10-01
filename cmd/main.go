package main

import (

	"github.com/spf13/cobra"
)

var (
	confFile string
)

func main() {
	serveCmd := &cobra.Command{
		Use: "serve",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: load config and run the server
		},
	}
	serveCmd.Flags().StringVar(&confFile, "config", "config.yaml", "The path to config file")
}
