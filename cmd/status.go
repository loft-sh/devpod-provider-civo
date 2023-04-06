package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/loft-sh/devpod-provider-civo/pkg/civo"

	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/provider"
	"github.com/spf13/cobra"
)

type InstanceStatus struct {
	NetworkInterfaces []InstanceStatusNetworkInterface `json:"networkInterfaces,omitempty"`
	Status            string                           `json:"status,omitempty"`
}

type InstanceStatusNetworkInterface struct {
	AccessConfigs []InstanceStatusAccessConfig `json:"accessConfigs,omitempty"`
}

type InstanceStatusAccessConfig struct {
	NatIP string `json:"natIP,omitempty"`
}

// StatusCmd holds the cmd flags
type StatusCmd struct{}

// NewStatusCmd defines a command
func NewStatusCmd() *cobra.Command {
	cmd := &StatusCmd{}
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Status an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			civoProvider, err := civo.NewProvider(log.Default)
			if err != nil {
				return err
			}

			return cmd.Run(
				context.Background(),
				civoProvider,
				provider.FromEnvironment(),
				log.Default,
			)
		},
	}

	return statusCmd
}

// Run runs the command logic
func (cmd *StatusCmd) Run(
	ctx context.Context,
	providerCivo *civo.CivoProvider,
	machine *provider.Machine,
	logs log.Logger,
) error {

	status, err := civo.Status(providerCivo)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, status)
	return err
}
