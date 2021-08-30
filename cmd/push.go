package cmd

import (
	"github.com/OLUWAMUYIWA/got/internal"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "pushes changes to a remote repository",
	Long:  `pushes changes to a remote repository`,
	Run: func(cmd *cobra.Command, args []string) {
		git := internal.NewGot()
		git.Push(args[0])
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)

}
