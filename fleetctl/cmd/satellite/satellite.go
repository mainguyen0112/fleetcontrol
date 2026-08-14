package satellite

import "github.com/spf13/cobra"

var SatelliteCmd = &cobra.Command{
	Use:   "satellite",
	Short: "Manage satellites (dev/debug only)",
}