package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Contract",
	Long:  `All software has versions. This is Contract's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(style.Render("Contract v0.1.0"))
	},
}
