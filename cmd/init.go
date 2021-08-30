package cmd

import (
	"github.com/OLUWAMUYIWA/got/internal"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the repository",
	Long: `Initializes the repository`,
	Run: func(cmd *cobra.Command, args []string) {
		got := internal.NewGot()
		got.Init(args[0])
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
