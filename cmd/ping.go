package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(ping)
}

var ping = &cobra.Command{
	Use:   "ping",
	Short: "reply pong",
	Long:  `This subcommand reply pong`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Pong !")
	},
}
