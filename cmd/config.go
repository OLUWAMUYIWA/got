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

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure user name and email",
	Long: `Saves the username and email to a config file on the computer. Whenever needed for commit, 
			got reaches for it and uses it. You can call this method more than once, as it replaces
			its existing content every time.`,
	Run: func(cmd *cobra.Command, args []string) {
		uname, err := cmd.Flags().GetString("uname")
		if err != nil {
			panic("provide the username")
		}
		email, err := cmd.Flags().GetString("email")
		if err != nil {
			panic("provide the email")
		}
		config := internal.ConfigObject{
			uname, 
			email,
		}
		err = internal.Config(config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().StringP("uname", "u", "", "takes the username")
	configCmd.Flags().StringP("email", "e", "", "takes the email address")
}
