package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List devices connected to the router",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := getSession()
		if err != nil {
			return err
		}
		devices, err := session.LocalMap()
		if err != nil {
			return fmt.Errorf("could not get device list: %w", err)
		}
		if len(devices) == 0 {
			fmt.Println("No devices found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tIP ADDRESS\tMAC ADDRESS\tCONNECTION\tOPEN PORTS")
		for _, d := range devices {
			openPorts := "no"
			if d.OpenPorts {
				openPorts = "yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				d.Name, d.IPAddress, d.MacAddress, d.ConnectionType, openPorts)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(devicesCmd)
}
