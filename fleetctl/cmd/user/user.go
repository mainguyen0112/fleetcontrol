package user

import "github.com/spf13/cobra"

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users (dev/debug only)",
}