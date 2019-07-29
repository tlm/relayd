package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "relayd",
	Short: "relayd grpc api for controlling state relays, specifically my garage door",
}

func init() {
	rootCmd.AddCommand(serveCommand)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
