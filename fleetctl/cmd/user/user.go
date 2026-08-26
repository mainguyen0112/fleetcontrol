package user

import (
	"fmt"

	"github.com/mainguyen0112/fleetcontrol/api/gen"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/output"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/runtime"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/secret"
	"github.com/spf13/cobra"
)

func NewCommand(opts *runtime.Options, passwords secret.PasswordReader) *cobra.Command {
	cmd := &cobra.Command{Use: "user", Short: "Manage users (dev/debug only)"}
	cmd.AddCommand(newCreateCommand(opts, passwords), newListCommand(opts))
	return cmd
}

func newCreateCommand(opts *runtime.Options, passwords secret.PasswordReader) *cobra.Command {
	var username, role string
	var passwordStdin bool
	cmd := &cobra.Command{Use: "create", Short: "Create a user (admin only)", RunE: func(cmd *cobra.Command, _ []string) error {
		if role != "admin" && role != "viewer" {
			return fmt.Errorf("role must be admin or viewer")
		}
		client, editors, err := opts.Client(true)
		if err != nil {
			return err
		}
		password, err := passwords.Read(cmd.InOrStdin(), cmd.ErrOrStderr(), passwordStdin)
		if err != nil {
			return err
		}
		apiRole := gen.CreateUserRequestRole(role)
		resp, err := client.PostUsersWithResponse(cmd.Context(), gen.CreateUserRequest{Username: username, Password: password, Role: &apiRole}, editors...)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		if err := output.StatusError(resp.StatusCode(), resp.Body, 201); err != nil {
			return err
		}
		return output.JSON(cmd.OutOrStdout(), resp.JSON201)
	}}
	cmd.Flags().StringVar(&username, "username", "", "username")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read password from standard input")
	cmd.Flags().StringVar(&role, "role", "viewer", "role (admin or viewer)")
	_ = cmd.MarkFlagRequired("username")
	return cmd
}

func newListCommand(opts *runtime.Options) *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List users (admin only)", RunE: func(cmd *cobra.Command, _ []string) error {
		client, editors, err := opts.Client(true)
		if err != nil {
			return err
		}
		resp, err := client.GetUsersWithResponse(cmd.Context(), editors...)
		if err != nil {
			return fmt.Errorf("list users: %w", err)
		}
		if err := output.StatusError(resp.StatusCode(), resp.Body, 200); err != nil {
			return err
		}
		return output.JSON(cmd.OutOrStdout(), resp.JSON200)
	}}
}
