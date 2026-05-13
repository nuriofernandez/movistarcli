package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/nuriofernandez/movistarapi/hgu"
	"github.com/spf13/cobra"
)

var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Manage port forwarding rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPorts()
	},
}

var portsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List port forwarding rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPorts()
	},
}

func listPorts() error {
	session, err := getSession()
	if err != nil {
		return err
	}
	ports, err := session.OpenPorts()
	if err != nil {
		return fmt.Errorf("could not get port rules: %w", err)
	}
	if len(ports) == 0 {
		fmt.Println("No port forwarding rules found.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPROTOCOL\tADDRESS\tEXT PORTS\tINT PORT\tINTERFACE\tENABLED")
	for _, p := range ports {
		extPorts := strconv.Itoa(p.ExternalPortStart)
		if p.ExternalPortEnd != p.ExternalPortStart {
			extPorts = fmt.Sprintf("%d-%d", p.ExternalPortStart, p.ExternalPortEnd)
		}
		enabled := "yes"
		if !p.Enabled {
			enabled = "no"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			p.Id, p.Name, p.Protocol, p.Address, extPorts, p.InternalPortStart, p.Interface, enabled)
	}
	return w.Flush()
}

// flags shared by add and update
var (
	portName      string
	portProtocol  string
	portAddress   string
	portExtStart  int
	portExtEnd    int
	portIntStart  int
	portEnabled   bool
	portInterface string
	portId        int
)

var portsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a port forwarding rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := getSession()
		if err != nil {
			return err
		}
		extEnd := portExtEnd
		if extEnd == 0 {
			extEnd = portExtStart
		}
		port := hgu.OpenPort{
			Name:              portName,
			Protocol:          hgu.Protocol(portProtocol),
			Address:           portAddress,
			ExternalPortStart: portExtStart,
			ExternalPortEnd:   extEnd,
			InternalPortStart: portIntStart,
			Enabled:           portEnabled,
			Interface:         portInterface,
		}
		if err := session.OpenPort(port); err != nil {
			return fmt.Errorf("could not add port rule: %w", err)
		}
		fmt.Printf("Port rule %q added.\n", portName)
		return nil
	},
}

var portsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing port forwarding rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := getSession()
		if err != nil {
			return err
		}
		extEnd := portExtEnd
		if extEnd == 0 {
			extEnd = portExtStart
		}
		port := hgu.OpenPort{
			Id:                portId,
			Name:              portName,
			Protocol:          hgu.Protocol(portProtocol),
			Address:           portAddress,
			ExternalPortStart: portExtStart,
			ExternalPortEnd:   extEnd,
			InternalPortStart: portIntStart,
			Enabled:           portEnabled,
			Interface:         portInterface,
		}
		if err := session.UpdatePort(port); err != nil {
			return fmt.Errorf("could not update port rule: %w", err)
		}
		fmt.Printf("Port rule %d updated.\n", portId)
		return nil
	},
}

var portsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a port forwarding rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id %q: must be a number", args[0])
		}
		session, err := getSession()
		if err != nil {
			return err
		}
		ports, err := session.OpenPorts()
		if err != nil {
			return fmt.Errorf("could not get port rules: %w", err)
		}
		var iface string
		for _, p := range ports {
			if p.Id == id {
				iface = p.Interface
				break
			}
		}
		if iface == "" {
			return fmt.Errorf("port rule %d not found", id)
		}
		if err := session.DeletePort(id, iface); err != nil {
			return fmt.Errorf("could not delete port rule: %w", err)
		}
		fmt.Printf("Port rule %d deleted.\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(portsCmd)
	portsCmd.AddCommand(portsListCmd, portsAddCmd, portsUpdateCmd, portsDeleteCmd)

	addPortFlags(portsAddCmd, false)
	addPortFlags(portsUpdateCmd, true)
}

func addPortFlags(cmd *cobra.Command, requireId bool) {
	if requireId {
		cmd.Flags().IntVar(&portId, "id", 0, "port rule ID to update")
		_ = cmd.MarkFlagRequired("id")
	}
	cmd.Flags().StringVar(&portName, "name", "", "rule name (max 16 chars)")
	cmd.Flags().StringVar(&portProtocol, "protocol", "TCP", "protocol: TCP, UDP, or BOTH")
	cmd.Flags().StringVar(&portAddress, "address", "", "target device IP address")
	cmd.Flags().IntVar(&portExtStart, "ext-start", 0, "external port range start")
	cmd.Flags().IntVar(&portExtEnd, "ext-end", 0, "external port range end (defaults to ext-start)")
	cmd.Flags().IntVar(&portIntStart, "int-start", 0, "internal port (maps from ext-start)")
	cmd.Flags().BoolVar(&portEnabled, "enabled", true, "enable the rule")
	cmd.Flags().StringVar(&portInterface, "interface", "ppp0.1", "WAN interface")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("address")
	_ = cmd.MarkFlagRequired("ext-start")
	_ = cmd.MarkFlagRequired("int-start")
}
