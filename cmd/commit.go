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

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "makes a commit to the local git",
	Long:  `makes a commit to the local git.`,
	Run: func(cmd *cobra.Command, args []string) {
		msg, _ := cmd.Flags().GetString("msg")
		git := internal.NewGit()
		git.Commit(msg)
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	addCmd.Flags().StringP("msg", "m", "new commit", "the message to be logged with the commit")
}
