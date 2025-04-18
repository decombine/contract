package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
	logoutCmd.Flags().StringP("network", "n", "decombine", "Network to logout from")
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from a Contract Network",
	Long:  `Logout from a Contract Network.`,
	RunE:  logoutExecute,
}

func logoutExecute(cmd *cobra.Command, args []string) error {

	name, err := cmd.Flags().GetString("network")
	if err != nil {
		return err
	}
	err = logout(name)
	if err != nil {
		return err
	}
	fmt.Println("Logged out from", name)
	return nil
}

func logout(name string) error {
	err := deleteToken(name)
	if err != nil {
		return err
	}
	return nil
}

func deleteToken(name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	tokenDir := home + `/.config/.` + name + `/`
	tokenPath := tokenDir + `token`
	err = os.Remove(tokenPath)
	if err != nil {
		return err
	}
	return nil
}
