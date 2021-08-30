package cmd

import (
	"fmt"
	"os"

	"github.com/OLUWAMUYIWA/got/internal"
	"github.com/spf13/cobra"
)

// hashObjectCmd represents the hashObject command
var hashObjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "Hash a given object, given the filename and type.",
	Long: `Hash a given object, given the filename and type.

Thye should be one of:
"blob",, "tree", "commit"
`,
	Run: func(cmd *cobra.Command, args []string) {
		std, err := cmd.Flags().GetBool("stdin")
		if err != nil {
			fmt.Printf("err: %s", err)
			os.Exit(1)
		}
		write, err := cmd.Flags().GetBool("write")
		if err != nil {
			fmt.Printf("err: %s", err)
			os.Exit(1)
		}
		if std {
			internal.HashFile("", write, std)
		} else {
			internal.HashFile(args[0], write, std)
		}
	},
}

func init() {
	rootCmd.AddCommand(hashObjectCmd)
	hashObjectCmd.Flags().BoolP("stdin", "s", false, "should I get the content from stdin?")
	hashObjectCmd.Flags().BoolP("write", "w", false, "write to the object directory?")
}
