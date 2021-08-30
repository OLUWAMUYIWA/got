package cmd

import (
	"github.com/OLUWAMUYIWA/got/internal"
	"github.com/spf13/cobra"
)

// catFileCmd represents the catFile command
var catFileCmd = &cobra.Command{
	Use:   "cat-file",
	Short: "displays the content of a file",
	Long: `displays the content of a file, given the name as positional arguments `,
	Run: func(cmd *cobra.Command, args []string) {
		got := internal.NewGot()
		// p, err := cmd.Flags().GetBool("pretty")
		// s, err := cmd.Flags().GetBool("size")
		// t, err := cmd.Flags().GetBool("type")
		// if err != nil {
		// 	fmt.Printf("shit happened: %s\n", err)
		// 	os.Exit(1)
		// }
		if (p && s) || (p && t) || (s && t) {
			panic("Only one of the flags should be on")
		}
		mode := ""
		//this ensures that default is pretty
		if (!s && !t) {
			mode = "pretty"
		}
		if s {
			mode = "size"
		} else if t {
			mode = "type"
		}
		got.CatFile(args[0], mode)
	},
}
var p, t, s bool

func init() {
	rootCmd.AddCommand(catFileCmd)
	catFileCmd.Flags().BoolVarP(&p, "pretty", "p", false, "choose what mode the file is created in, pretty?")
	catFileCmd.Flags().BoolVarP(&t, "type", "t", false, "choose what mode the file is created in, type?")
	catFileCmd.Flags().BoolVarP(&s, "size", "s", false, "choose what mode the file is created in, size?")
}
