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
		got.LsFiles(Details)
	},
}
var Details bool

func init() {
	rootCmd.AddCommand(lsFilesCmd)
	addCmd.Flags().BoolVarP(&Details, "details", "d", false, "detailsshould we show deails?")
}
