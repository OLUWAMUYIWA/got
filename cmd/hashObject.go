/*
Copyright © 2021 Oluwamuyiwa Onigbinde <onigbindemy@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
			internal.HashFile("", args[1], write, std)
		} else {
			internal.HashFile(args[0], args[1], write, std)
		}
	},
}

func init() {
	rootCmd.AddCommand(hashObjectCmd)
	hashObjectCmd.Flags().BoolP("stdin", "s", false, "should I get the content from stdin?")
	hashObjectCmd.Flags().BoolP("write", "w", false, "write to the object directory?")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hashObjectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hashObjectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
