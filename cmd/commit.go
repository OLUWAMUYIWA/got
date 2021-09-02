package cmd

import (
	"os"

	"github.com/OLUWAMUYIWA/got/internal"
	"github.com/spf13/cobra"
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "makes a commit to the local git",
	Long:  `makes a commit to the local git.`,
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := cmd.Flags().GetString("msg")
		if err != nil {
			os.Stdout.WriteString("Got err: " + err.Error())
			os.Exit(1)
		}
		git := internal.NewGot()
		git.Commit(msg)
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().StringP("msg", "m", "new commit", "the message to be logged with the commit")
}
