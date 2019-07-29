package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tlmiller/relayd/pkg/api"
)

var serveCommand = &cobra.Command{
	Use:   "serve",
	Short: "start the relayd api server",
	Run:   serve,
}

func serve(_ *cobra.Command, _ []string) {
	server, err := api.NewServer()
	if err != nil {
		fmt.Printf("error starting new server %v", err)
		os.Exit(1)
	}

	if err := server.Serve(); err != nil {
		fmt.Printf("error serving api %v", err)
		os.Exit(1)
	}
}
