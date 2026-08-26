package cmd

import (
	"fmt"
	"os"

	"github.com/mainguyen0112/fleetcontrol/api/gen"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/cmd/satellite"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/cmd/user"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/output"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/runtime"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/secret"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	return newRootCommand(secret.NewPasswordReader())
}

func newRootCommand(passwords secret.PasswordReader) *cobra.Command {
	opts := &runtime.Options{}
	rootCmd := &cobra.Command{
		Use:           "fleetctl",
		Short:         "fleetctl is a dev/debug CLI for FleetControl",
		SilenceErrors: true,
		SilenceUsage:  true,
		Long: `fleetctl is a development and debugging tool for FleetControl.

It is NOT the official way to manage production fleet state — GitOps
via ArgoCD and Kubernetes CRDs is the source of truth. fleetctl exists
for local development, testing, and debugging only.`,
	}
	rootCmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", "config file (default ~/.fleetctl/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&opts.Server, "server", "", "FleetControl API URL (overrides config)")
	rootCmd.AddCommand(newLoginCommand(opts, passwords), newHealthCommand(opts), satellite.NewCommand(opts), user.NewCommand(opts, passwords))
	return rootCmd
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newLoginCommand(opts *runtime.Options, passwords secret.PasswordReader) *cobra.Command {
	var username string
	var passwordStdin bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate and save the development token",
		RunE: func(cmd *cobra.Command, _ []string) error {
			password, err := passwords.Read(cmd.InOrStdin(), cmd.ErrOrStderr(), passwordStdin)
			if err != nil {
				return err
			}
			client, _, err := opts.Client(false)
			if err != nil {
				return err
			}
			resp, err := client.PostAuthLoginWithResponse(cmd.Context(), gen.LoginRequest{Username: username, Password: password})
			if err != nil {
				return fmt.Errorf("login request: %w", err)
			}
			if err := output.StatusError(resp.StatusCode(), resp.Body, 200); err != nil {
				return err
			}
			if resp.JSON200 == nil || resp.JSON200.Token == nil {
				return fmt.Errorf("login response did not contain a token")
			}
			cfg, err := opts.Load()
			if err != nil {
				return err
			}
			cfg.Token = *resp.JSON200.Token
			if err := opts.Save(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Login successful")
			return nil
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "username")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read password from standard input")
	_ = cmd.MarkFlagRequired("username")
	return cmd
}

func newHealthCommand(opts *runtime.Options) *cobra.Command {
	return &cobra.Command{
		Use: "health", Short: "Check Control Plane health",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, editors, err := opts.Client(false)
			if err != nil {
				return err
			}
			resp, err := client.GetHealthWithResponse(cmd.Context(), editors...)
			if err != nil {
				return fmt.Errorf("health request: %w", err)
			}
			if err := output.StatusError(resp.StatusCode(), resp.Body, 200); err != nil {
				return err
			}
			return output.JSON(cmd.OutOrStdout(), resp.JSON200)
		},
	}
}
