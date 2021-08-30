package cmd

import (
	"github.com/OLUWAMUYIWA/got/internal"
	"github.com/spf13/cobra"
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "makes a commit to the local git",
	Long:  `makes a commit to the local git.`,
	Run: func(cmd *cobra.Command, args []string) {
		msg, _ := cmd.Flags().GetString("msg")
		git := internal.NewGot()
		git.Commit(msg)
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	addCmd.Flags().StringP("msg", "m", "new commit", "the message to be logged with the commit")
}
