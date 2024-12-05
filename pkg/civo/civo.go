package civo

import (
	"encoding/json"
	"os"

	"github.com/loft-sh/devpod-provider-civo/pkg/options"
	"github.com/loft-sh/devpod/pkg/client"
	"github.com/loft-sh/devpod/pkg/log"

	"github.com/civo/civogo"
	"github.com/pkg/errors"
)

type CivoToken struct {
	APIKey string "json:apikey"
	Region string "json:region"
}

var tokenJSON CivoToken

func NewProvider(withFolder bool, logs log.Logger) (*CivoProvider, error) {
	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		err := json.Unmarshal([]byte(civoToken), &tokenJSON)
		if err != nil {
			return nil, err
		}

		err = os.Setenv("CIVO_API_KEY", tokenJSON.APIKey)
		if err != nil {
			return nil, err
		}

		err = os.Setenv("CIVO_REGION", tokenJSON.Region)
		if err != nil {
			return nil, err
		}
	}

	civoApiKey := os.Getenv("CIVO_API_KEY")
	if civoApiKey == "" {
		return nil, errors.Errorf("CIVO_API_KEY is not set")
	}

	civoRegion := os.Getenv("CIVO_REGION")
	if civoRegion == "" {
		return nil, errors.Errorf("CIVO_REGION is not set")
	}

	config, err := options.FromEnv(false, withFolder)

	if err != nil {
		return nil, err
	}

	client, err := civogo.NewClient(civoApiKey, civoRegion)
	if err != nil {
		return nil, err
	}

	// create provider
	provider := &CivoProvider{
		Config: config,
		Client: client,
		Log:    logs,
	}

	return provider, nil
}

type CivoProvider struct {
	Config           *options.Options
	Client           *civogo.Client
	Log              log.Logger
	WorkingDirectory string
}

func AccessToken() (string, error) {
	// If the user is logged via token, just forward it
	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		return civoToken, nil
	}

	civoApiKey := os.Getenv("CIVO_API_KEY")
	if civoApiKey == "" {
		return "", errors.Errorf("CIVO_API_KEY is not set")
	}

	civoRegion := os.Getenv("CIVO_REGION")
	if civoRegion == "" {
		return "", errors.Errorf("CIVO_REGION is not set")
	}

	tokenJSON.APIKey = civoApiKey
	tokenJSON.Region = civoRegion

	result, err := json.Marshal(tokenJSON)

	return string(result), err
}

func GetDevpodInstance(civoProvider *CivoProvider) (*civogo.Instance, error) {
	return civoProvider.Client.FindInstance(civoProvider.Config.MachineID)
}

func Create(civoProvider *CivoProvider) error {

	config, err := civoProvider.Client.NewInstanceConfig()
	if err != nil {
		return err
	}

	config.PublicIPRequired = "true"
	config.Count = 1
	config.Hostname = civoProvider.Config.MachineID
	config.Size = civoProvider.Config.MachineType
	config.Region = civoProvider.Config.Region
	config.PublicIPRequired = "true"

	_, err = civoProvider.Client.CreateInstance(config)
	if err != nil {
		return err
	}

	return nil
}
func Delete(civoProvider *CivoProvider) error {
	instance, err := GetDevpodInstance(civoProvider)
	if err != nil {
		return err
	}

	_, err = civoProvider.Client.DeleteInstance(instance.ID)
	if err != nil {
		return err
	}

	return nil
}

func Start(civoProvider *CivoProvider) error {
	instance, err := GetDevpodInstance(civoProvider)
	if err != nil {
		return err
	}

	_, err = civoProvider.Client.StartInstance(instance.ID)
	if err != nil {
		return err
	}

	return nil
}

func Stop(civoProvider *CivoProvider) error {
	instance, err := GetDevpodInstance(civoProvider)
	if err != nil {
		return err
	}

	_, err = civoProvider.Client.StopInstance(instance.ID)
	if err != nil {
		return err
	}

	return nil
}

func Status(civoProvider *CivoProvider) (client.Status, error) {
	instance, err := GetDevpodInstance(civoProvider)
	if err != nil {
		return client.StatusNotFound, nil
	}

	switch {
	case instance.Status == "ACTIVE":
		return client.StatusRunning, nil
	case instance.Status == "SHUTOFF":
		return client.StatusStopped, nil
	default:
		return client.StatusBusy, nil
	}
}
