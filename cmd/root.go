package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	movistarapi "github.com/nuriofernandez/movistarapi"
	"github.com/nuriofernandez/movistarapi/hgu"
	"github.com/spf13/cobra"
)

var password string

var rootCmd = &cobra.Command{
	Use:          "movistar",
	Short:        "CLI for Movistar HGU router management",
	Long:         "Control and configure your Movistar HGU router from the command line.",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "router password (or set MOVISTAR_PASSWORD env var)")
}

const credentialsFile = ".config/Movistar/credentials"

func readCredentialsFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	f, err := os.Open(filepath.Join(home, credentialsFile))
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(key) == "password" {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func getSession() (hgu.HGUSession, error) {
	pass := password
	if pass == "" {
		pass = os.Getenv("MOVISTAR_PASSWORD")
	}
	if pass == "" {
		pass = readCredentialsFile()
	}
	if pass == "" {
		return hgu.HGUSession{}, fmt.Errorf(
			"password required: use --password, set MOVISTAR_PASSWORD, or add it as 'password=<pass>' to ~/%s", credentialsFile)
	}
	session, err := movistarapi.HGULogin(pass)
	if err != nil {
		return hgu.HGUSession{}, fmt.Errorf("login failed: %w", err)
	}
	if !session.IsValid {
		return hgu.HGUSession{}, fmt.Errorf("login failed: invalid session (wrong password?)")
	}
	return session, nil
}
