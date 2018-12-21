package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of app",
	Long:  `All software has versions. This is app's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("app -- HEAD")
	},
}
