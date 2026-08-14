package cmd

import (
	"fmt"
	"os"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/cmd/satellite"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/cmd/user"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fleetctl",
	Short: "fleetctl is a dev/debug CLI for FleetControl",
	Long: `fleetctl is a development and debugging tool for FleetControl.

It is NOT the official way to manage production fleet state — GitOps
via ArgoCD and Kubernetes CRDs is the source of truth. fleetctl exists
for local development, testing, and debugging only.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(satellite.SatelliteCmd)
	rootCmd.AddCommand(user.UserCmd)
}