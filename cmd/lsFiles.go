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
		det, _ := cmd.Flags().GetBool("details")
		got.LsFiles(det)
	},
}

func init() {
	rootCmd.AddCommand(lsFilesCmd)
	addCmd.Flags().BoolP("details", "d", true, "should we show deails?")
}
