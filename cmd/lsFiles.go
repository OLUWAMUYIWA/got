package cmd

import (
	"github.com/OLUWAMUYIWA/got/internal"
	"github.com/spf13/cobra"
)

// lsFilesCmd represents the lsFiles command
var lsFilesCmd = &cobra.Command{
	Use:   "ls-files",
	Short: "List files",
	Long: ` It displays the mode sha1, path and type of object files to stdout 
	`,
	Run: func(cmd *cobra.Command, args []string) {
		got := internal.NewGot()
		got.LsFiles(Stage)
	},
}
var Stage bool

func init() {
	rootCmd.AddCommand(lsFilesCmd)
	addCmd.Flags().BoolVarP(&Stage, "stage", "s", false, "should we show deails?, without this flag, only file names will be shown")
}
