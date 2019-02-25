package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	confFile string
)

var rootCmd = &cobra.Command{
	Use:   "serve",
	Short: "Hugo is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		// TODO: load config and run the server
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
