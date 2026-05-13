package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/nuriofernandez/movistarapi/hgu"
	"github.com/spf13/cobra"
)

var (
	colorTCP  = color.New(color.FgCyan).SprintFunc()
	colorUDP  = color.New(color.FgYellow).SprintFunc()
	colorBOTH = color.New(color.FgMagenta).SprintFunc()
)

func colorizeProtocol(line string) string {
	line = strings.ReplaceAll(line, "TCP", colorTCP("TCP"))
	line = strings.ReplaceAll(line, "UDP", colorUDP("UDP"))
	line = strings.ReplaceAll(line, "BOTH", colorBOTH("BOTH"))
	return line
}

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
	printTable("ID\tNAME\tPROTOCOL\tADDRESS\tEXT PORTS\tINT PORT\tINTERFACE\tENABLED", func(w *tabwriter.Writer) {
		for _, p := range ports {
			extPorts := strconv.Itoa(p.ExternalPortStart)
			if p.ExternalPortEnd != p.ExternalPortStart {
				extPorts = fmt.Sprintf("%d-%d", p.ExternalPortStart, p.ExternalPortEnd)
			}
			enabled := "no"
			if p.Enabled {
				enabled = "yes"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
				p.Id, p.Name, p.Protocol, p.Address, extPorts, p.InternalPortStart, p.Interface, enabled)
		}
	}, colorizeProtocol)
	return nil
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
		intStart := portIntStart
		if intStart == 0 {
			intStart = portExtStart
		}
		port := hgu.OpenPort{
			Name:              portName,
			Protocol:          hgu.Protocol(portProtocol),
			Address:           portAddress,
			ExternalPortStart: portExtStart,
			ExternalPortEnd:   extEnd,
			InternalPortStart: intStart,
			Enabled:           portEnabled,
			Interface:         portInterface,
		}
		if err := session.OpenPort(port); err != nil {
			return fmt.Errorf("could not add port rule: %w", err)
		}
		color.Green("Port rule %q added.", portName)
		return nil
	},
}

var portsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an existing port forwarding rule",
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
		var port *hgu.OpenPort
		for i := range ports {
			if ports[i].Id == id {
				port = &ports[i]
				break
			}
		}
		if port == nil {
			return fmt.Errorf("port rule %d not found", id)
		}
		if cmd.Flags().Changed("name") {
			port.Name = portName
		}
		if cmd.Flags().Changed("protocol") {
			port.Protocol = hgu.Protocol(portProtocol)
		}
		if cmd.Flags().Changed("address") {
			port.Address = portAddress
		}
		if cmd.Flags().Changed("ext-start") {
			original := port.ExternalPortStart
			port.ExternalPortStart = portExtStart
			if !cmd.Flags().Changed("ext-end") && port.ExternalPortEnd == original {
				port.ExternalPortEnd = portExtStart
			}
			if !cmd.Flags().Changed("int-start") && port.InternalPortStart == original {
				port.InternalPortStart = portExtStart
			}
		}
		if cmd.Flags().Changed("ext-end") {
			port.ExternalPortEnd = portExtEnd
		}
		if cmd.Flags().Changed("int-start") {
			port.InternalPortStart = portIntStart
		}
		if cmd.Flags().Changed("enabled") {
			port.Enabled = portEnabled
		}
		if cmd.Flags().Changed("interface") {
			port.Interface = portInterface
		}
		if err := session.UpdatePort(*port); err != nil {
			return fmt.Errorf("could not update port rule: %w", err)
		}
		color.Green("Port rule %d updated.", id)
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
		color.Green("Port rule %d deleted.", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(portsCmd)
	portsCmd.AddCommand(portsListCmd, portsAddCmd, portsUpdateCmd, portsDeleteCmd, portsEnableCmd, portsDisableCmd)

	addPortFlags(portsAddCmd)
	addPortFlags(portsUpdateCmd)
}

func setEnabled(id int, enabled bool) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	ports, err := session.OpenPorts()
	if err != nil {
		return fmt.Errorf("could not get port rules: %w", err)
	}
	var port *hgu.OpenPort
	for i := range ports {
		if ports[i].Id == id {
			port = &ports[i]
			break
		}
	}
	if port == nil {
		return fmt.Errorf("port rule %d not found", id)
	}
	port.Enabled = enabled
	return session.UpdatePort(*port)
}

var portsEnableCmd = &cobra.Command{
	Use:   "enable <id>",
	Short: "Enable a port forwarding rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id %q: must be a number", args[0])
		}
		if err := setEnabled(id, true); err != nil {
			return err
		}
		color.Green("Port rule %d enabled.", id)
		return nil
	},
}

var portsDisableCmd = &cobra.Command{
	Use:   "disable <id>",
	Short: "Disable a port forwarding rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid id %q: must be a number", args[0])
		}
		if err := setEnabled(id, false); err != nil {
			return err
		}
		color.Yellow("Port rule %d disabled.", id)
		return nil
	},
}

func addPortFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&portName, "name", "", "rule name (max 16 chars)")
	cmd.Flags().StringVar(&portProtocol, "protocol", "", "protocol: TCP, UDP, or BOTH")
	cmd.Flags().StringVar(&portAddress, "address", "", "target device IP address")
	cmd.Flags().IntVar(&portExtStart, "ext-start", 0, "external port range start")
	cmd.Flags().IntVar(&portExtEnd, "ext-end", 0, "external port range end (defaults to ext-start)")
	cmd.Flags().IntVar(&portIntStart, "int-start", 0, "internal port (maps from ext-start)")
	cmd.Flags().BoolVar(&portEnabled, "enabled", true, "enable the rule")
	cmd.Flags().StringVar(&portInterface, "interface", "ppp0.1", "WAN interface")
	if cmd == portsAddCmd {
		_ = cmd.MarkFlagRequired("name")
		_ = cmd.MarkFlagRequired("protocol")
		_ = cmd.MarkFlagRequired("address")
		_ = cmd.MarkFlagRequired("ext-start")
	}
}
