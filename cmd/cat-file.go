/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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

// catFileCmd represents the catFile command
var catFileCmd = &cobra.Command{
	Use:   "cat-file",
	Short: "displays the content of a file",
	Long: `displays the content of a file, given the name as positional arguments `,
	Run: func(cmd *cobra.Command, args []string) {
		got := internal.NewGot()
		p, err := cmd.Flags().GetBool("pretty")
		s, err := cmd.Flags().GetBool("size")
		t, err := cmd.Flags().GetBool("type")
		if err != nil {
			fmt.Printf("shit happened: %s\n", err)
			os.Exit(1)
		}
		if (p && s) || (p && t) || (s && t) {
			panic("Only one of the flags should be on")
		}
		mode := ""
		//this ensures that default is pretty
		if (!p && !s && !t) {
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

func init() {
	rootCmd.AddCommand(catFileCmd)
	rootCmd.Flags().BoolP("pretty", "p", false, "choose what mode the file is created in, pretty?")
	rootCmd.Flags().BoolP("type", "t", false, "choose what mode the file is created in, type?")
	rootCmd.Flags().BoolP("size", "s", false, "choose what mode the file is created in, size?")
}
