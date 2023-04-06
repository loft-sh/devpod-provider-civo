package cmd

import (
	"context"
	"fmt"

	"github.com/loft-sh/devpod-provider-civo/pkg/civo"

	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/provider"
	"github.com/spf13/cobra"
)

type InstanceToken struct {
	NetworkInterfaces []InstanceTokenNetworkInterface `json:"networkInterfaces,omitempty"`
	Token             string                          `json:"token,omitempty"`
}

type InstanceTokenNetworkInterface struct {
	AccessConfigs []InstanceTokenAccessConfig `json:"accessConfigs,omitempty"`
}

type InstanceTokenAccessConfig struct {
	NatIP string `json:"natIP,omitempty"`
}

// TokenCmd holds the cmd flags
type TokenCmd struct{}

// NewTokenCmd defines a command
func NewTokenCmd() *cobra.Command {
	cmd := &TokenCmd{}
	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Token an instance",
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

	return tokenCmd
}

// Run runs the command logic
func (cmd *TokenCmd) Run(
	ctx context.Context,
	providerCivo *civo.CivoProvider,
	machine *provider.Machine,
	logs log.Logger,
) error {
	token, err := civo.AccessToken()
	if err != nil {
		return err
	}

	fmt.Println(token)
	return nil
}
