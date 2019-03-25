package cmd

import (
	"fmt"
	"os"
    "os/signal"
    "syscall"
    "context"

	"github.com/remones/gsocks/config"
	"github.com/remones/gsocks/proxy"
	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	cfgFile string
	version = "v0.1"
)

var (
	rootCmd = &cobra.Command{
		Use: "help",
	}
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "start a gsocks server",
		Long:  `start a gsocks sever`,
		Run: func(cmd *cobra.Command, args []string) {
			srv := proxy.NewServer(cfg)
            idleConnsClosed := make(chan struct{})

            go func() {
                sigCh := make(chan os.Signal, 1)
                signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
                <-sigCh

                if err := srv.Close(context.Background()); err != nil {
                    fmt.Printf("gsocks server Shutdown: %v", err)
                }
                close(idleConnsClosed)
            }()

			if err := srv.ListenAndServe(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
            <-idleConnsClosed
		},
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of gsocks",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if cfgFile != "" && cfg == nil {
		cfg = config.NewConfig()
		if err := cfg.Load(cfgFile); err != nil {
			fmt.Println("Can't read config file: ", err)
			os.Exit(1)
		}
	}
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
