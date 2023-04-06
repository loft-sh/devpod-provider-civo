package cmd

import (
	"context"
	"os"

	"github.com/civo/civogo"
	"github.com/loft-sh/devpod-provider-civo/pkg/options"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/provider"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// InitCmd holds the cmd flags
type InitCmd struct{}

// NewInitCmd defines a init
func NewInitCmd() *cobra.Command {
	cmd := &InitCmd{}
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Init account",
		RunE: func(_ *cobra.Command, args []string) error {

			return cmd.Run(
				context.Background(),
				provider.FromEnvironment(),
				log.Default,
			)
		},
	}

	return initCmd
}

// Run runs the init logic
func (cmd *InitCmd) Run(
	ctx context.Context,
	machine *provider.Machine,
	logs log.Logger,
) error {
	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken == "" {
		return errors.Errorf("CIVO_TOKEN is not set")
	}

	civoRegion := os.Getenv("CIVO_REGION")
	if civoRegion == "" {
		return errors.Errorf("CIVO_REGION is not set")
	}

	_, err := options.FromEnv(true)

	if err != nil {
		return err
	}

	_, err = civogo.NewClient(civoToken, civoRegion)
	if err != nil {
		return err
	}

	return nil
}
