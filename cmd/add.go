package cmd

import (
	"github.com/OLUWAMUYIWA/got/internal"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a series of files to be staged for commit",
	Long:  `Adds a series of files to be staged for commit.`,
	Run: func(cmd *cobra.Command, args []string) {
		got := internal.NewGot()
		got.Add(args)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
