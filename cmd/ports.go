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
	printTable("ID\tNAME\tPROTOCOL\tADDRESS\tINT PORT\tEXT PORTS\tINTERFACE\tENABLED", func(w *tabwriter.Writer) {
		for _, p := range ports {
			extPorts := strconv.Itoa(p.ExternalPortStart)
			intPorts := strconv.Itoa(p.InternalPortStart)
			if p.ExternalPortEnd != 0 && p.ExternalPortEnd != p.ExternalPortStart {
				extPorts = fmt.Sprintf("%d:%d", p.ExternalPortStart, p.ExternalPortEnd)
				intPorts = fmt.Sprintf("%d:%d", p.InternalPortStart, p.InternalPortStart+(p.ExternalPortEnd-p.ExternalPortStart))
			}
			enabled := "no"
			if p.Enabled {
				enabled = "yes"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				p.Id, p.Name, p.Protocol, p.Address, intPorts, extPorts, p.Interface, enabled)
		}
	}, colorizeProtocol)
	return nil
}

// flags shared by add and update
var (
	portName      string
	portProtocol  string
	portAddress   string
	portSpec      string // -p format: 80 | 80:81 | 80:81/90 | 80:81/90-91
	portEnabled   bool
	portInterface string
)

// parsePortSpec parses the -p flag. Internal is left, external is right.
//
//	80          → int=80,    ext=80
//	80:81       → int=80-81, ext=80-81
//	80/90       → int=80,    ext=90
//	80:81/90    → int=80-81, ext=90-91 (ext end derived from range size)
//	80:81/90:91 → int=80-81, ext=90-91 (also accepts 90-91)
func parsePortSpec(s string) (extStart, extEnd, intStart int, err error) {
	parseRange := func(part string) (lo, hi int, e error) {
		startStr, endStr, hasRange := strings.Cut(part, ":")
		if !hasRange {
			startStr, endStr, hasRange = strings.Cut(part, "-")
		}
		lo, e = strconv.Atoi(startStr)
		if e != nil {
			return 0, 0, fmt.Errorf("invalid port %q", startStr)
		}
		hi = lo
		if hasRange {
			hi, e = strconv.Atoi(endStr)
			if e != nil {
				return 0, 0, fmt.Errorf("invalid port %q", endStr)
			}
		}
		return lo, hi, nil
	}

	intPart, extPart, hasSlash := strings.Cut(s, "/")

	intLo, intHi, err := parseRange(intPart)
	if err != nil {
		return 0, 0, 0, err
	}

	if !hasSlash {
		return intLo, intHi, intLo, nil
	}

	extLo, extHi, err := parseRange(extPart)
	if err != nil {
		return 0, 0, 0, err
	}

	// single ext port with an int range → derive ext end
	if !strings.ContainsAny(extPart, ":-") {
		extHi = extLo + (intHi - intLo)
	} else if extHi-extLo != intHi-intLo {
		return 0, 0, 0, fmt.Errorf("internal range size (%d) does not match external range size (%d)", intHi-intLo+1, extHi-extLo+1)
	}

	return extLo, extHi, intLo, nil
}

var portsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a port forwarding rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := getSession()
		if err != nil {
			return err
		}
		extStart, extEnd, intStart, err := parsePortSpec(portSpec)
		if err != nil {
			return err
		}
		port := hgu.OpenPort{
			Name:              portName,
			Protocol:          hgu.Protocol(portProtocol),
			Address:           portAddress,
			ExternalPortStart: extStart,
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
		if cmd.Flags().Changed("port") {
			extStart, extEnd, intStart, err := parsePortSpec(portSpec)
			if err != nil {
				return err
			}
			port.ExternalPortStart = extStart
			port.ExternalPortEnd = extEnd
			port.InternalPortStart = intStart
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
	cmd.Flags().StringVarP(&portSpec, "port", "p", "", "port spec: 80 | 80:81 | 80/90 | 80:81/90:91  (internal/external)")
	cmd.Flags().BoolVar(&portEnabled, "enabled", true, "enable the rule")
	cmd.Flags().StringVar(&portInterface, "interface", "ppp0.1", "WAN interface")
	if cmd == portsAddCmd {
		_ = cmd.MarkFlagRequired("name")
		_ = cmd.MarkFlagRequired("protocol")
		_ = cmd.MarkFlagRequired("address")
		_ = cmd.MarkFlagRequired("port")
	}
}
