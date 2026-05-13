package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboot the router",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := getSession()
		if err != nil {
			return err
		}
		if err := session.Reboot(); err != nil {
			return fmt.Errorf("reboot failed: %w", err)
		}
		fmt.Println("Router is rebooting...")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rebootCmd)
}
