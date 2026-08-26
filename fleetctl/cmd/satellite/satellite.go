package satellite

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mainguyen0112/fleetcontrol/api/gen"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/output"
	"github.com/mainguyen0112/fleetcontrol/fleetctl/internal/runtime"
	"github.com/spf13/cobra"
)

func NewCommand(opts *runtime.Options) *cobra.Command {
	cmd := &cobra.Command{Use: "satellite", Short: "Manage satellites (dev/debug only)"}
	cmd.AddCommand(newCreateCommand(opts), newGetCommand(opts), newListCommand(opts), newDeleteCommand(opts))
	return cmd
}

func newCreateCommand(opts *runtime.Options) *cobra.Command {
	var name, region string
	cmd := &cobra.Command{Use: "create", Short: "Create a manually managed satellite", RunE: func(cmd *cobra.Command, _ []string) error {
		client, editors, err := opts.Client(true)
		if err != nil {
			return err
		}
		resp, err := client.PostSatellitesWithResponse(cmd.Context(), gen.CreateSatelliteRequest{Name: name, Region: region}, editors...)
		if err != nil {
			return fmt.Errorf("create satellite: %w", err)
		}
		if err := output.StatusError(resp.StatusCode(), resp.Body, 201); err != nil {
			return err
		}
		return output.JSON(cmd.OutOrStdout(), resp.JSON201)
	}}
	cmd.Flags().StringVar(&name, "name", "", "satellite name")
	cmd.Flags().StringVar(&region, "region", "", "satellite region")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("region")
	return cmd
}

func newGetCommand(opts *runtime.Options) *cobra.Command {
	return &cobra.Command{Use: "get <id>", Short: "Get a satellite", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid satellite ID: %w", err)
		}
		client, editors, err := opts.Client(true)
		if err != nil {
			return err
		}
		resp, err := client.GetSatellitesIdWithResponse(cmd.Context(), id, editors...)
		if err != nil {
			return fmt.Errorf("get satellite: %w", err)
		}
		if err := output.StatusError(resp.StatusCode(), resp.Body, 200); err != nil {
			return err
		}
		return output.JSON(cmd.OutOrStdout(), resp.JSON200)
	}}
}

func newListCommand(opts *runtime.Options) *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List satellites", RunE: func(cmd *cobra.Command, _ []string) error {
		client, editors, err := opts.Client(true)
		if err != nil {
			return err
		}
		resp, err := client.GetSatellitesWithResponse(cmd.Context(), editors...)
		if err != nil {
			return fmt.Errorf("list satellites: %w", err)
		}
		if err := output.StatusError(resp.StatusCode(), resp.Body, 200); err != nil {
			return err
		}
		return output.JSON(cmd.OutOrStdout(), resp.JSON200)
	}}
}

func newDeleteCommand(opts *runtime.Options) *cobra.Command {
	return &cobra.Command{Use: "delete <id>", Short: "Delete a manually managed satellite", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid satellite ID: %w", err)
		}
		client, editors, err := opts.Client(true)
		if err != nil {
			return err
		}
		getResp, err := client.GetSatellitesIdWithResponse(cmd.Context(), id, editors...)
		if err != nil {
			return fmt.Errorf("check satellite ownership: %w", err)
		}
		if err := output.StatusError(getResp.StatusCode(), getResp.Body, 200); err != nil {
			return err
		}
		if getResp.JSON200 != nil && getResp.JSON200.ManagedBy != nil && *getResp.JSON200.ManagedBy == gen.Operator {
			return fmt.Errorf("satellite is managed by the Operator; update its CR in Git instead")
		}
		resp, err := client.DeleteSatellitesIdWithResponse(cmd.Context(), id, editors...)
		if err != nil {
			return fmt.Errorf("delete satellite: %w", err)
		}
		if err := output.StatusError(resp.StatusCode(), resp.Body, 204); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Satellite deleted")
		return nil
	}}
}
