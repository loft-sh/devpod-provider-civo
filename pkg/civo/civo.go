package civo

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/loft-sh/devpod-provider-civo/pkg/options"
	"github.com/loft-sh/devpod/pkg/client"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/ssh"

	"github.com/civo/civogo"
	"github.com/pkg/errors"
)

type CivoToken struct {
	APIKey string "json:apikey"
	Region string "json:region"
}

var tokenJSON CivoToken

func NewProvider(logs log.Logger) (*CivoProvider, error) {
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

	config, err := options.FromEnv(false)

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

func GetInjectKeypairScript(civoProvider *CivoProvider, volume *civogo.Volume) (string, error) {
	publicKeyBase, err := ssh.GetPublicKeyBase(civoProvider.Config.MachineFolder)
	if err != nil {
		return "", err
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return "", err
	}

	resultScript := `#!/bin/sh
mkdir -p /home/devpod
# If disk is unformatted, let's format it first
if [ "ext4" != "$(blkid -o value -s TYPE /dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_` + volume.ID + `)" ]; then
	mkfs.ext4 /dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_` + volume.ID + `
fi
mount -o discard,defaults,noatime /dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_` + volume.ID + ` /home/devpod

# Move docker data dir
if command -v docker; then
	service docker stop
fi
mkdir -p /etc/docker
cat > /etc/docker/daemon.json << EOF
{
  "data-root": "/home/devpod/.docker-daemon",
  "live-restore": true
}
EOF

chattr +i /etc/docker/daemon.json
chattr +i /etc/docker/

# Make sure we only copy if volumes isn't initialized
if [ ! -d "/home/devpod/.docker-daemon" ]; then
  mkdir -p /home/devpod/.docker-daemon
  rsync -aP /var/lib/docker/ /home/devpod/.docker-daemon
fi

if command -v docker; then
	service docker start
fi

useradd devpod -d /home/devpod
mkdir -p /home/devpod
if grep -q sudo /etc/groups; then
	usermod -aG sudo devpod
elif grep -q wheel /etc/groups; then
	usermod -aG wheel devpod
fi
echo "devpod ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-devpod
mkdir -p /home/devpod/.ssh
echo "` + string(publicKey) + `" >> /home/devpod/.ssh/authorized_keys
chmod 0700 /home/devpod/.ssh
chmod 0600 /home/devpod/.ssh/authorized_keys
chown -R devpod:devpod /home/devpod`

	return resultScript, nil
}

func GetDevpodInstance(civoProvider *CivoProvider) (*civogo.Instance, error) {
	return civoProvider.Client.FindInstance(civoProvider.Config.MachineID)
}

// CreateOrReturnDisk will return the instance disk if exists, and create it
// if it doesn't
func CreateVolume(civoProvider *CivoProvider) (*civogo.Volume, error) {
	network, err := civoProvider.Client.GetDefaultNetwork()
	if err != nil {
		return nil, err
	}

	volumeConfig := civogo.VolumeConfig{
		Name:          civoProvider.Config.MachineID,
		NetworkID:     network.ID,
		Region:        civoProvider.Config.Region,
		SizeGigabytes: civoProvider.Config.DiskSizeGB,
		Bootable:      false,
	}

	_, err = civoProvider.Client.NewVolume(&volumeConfig)
	if err != nil {
		return nil, err
	}

	return GetVolume(civoProvider)
}

func GetVolume(civoProvider *CivoProvider) (*civogo.Volume, error) {
	return civoProvider.Client.FindVolume(civoProvider.Config.MachineID)
}

func CreateOrStart(civoProvider *CivoProvider, start bool) error {

	config, err := civoProvider.Client.NewInstanceConfig()
	if err != nil {
		return err
	}

	var volume *civogo.Volume

	// if it's a start, we already have the volume
	if start {
		volume, err = GetVolume(civoProvider)
		if err != nil {
			return err
		}

	} else {
		volume, err = CreateVolume(civoProvider)
		if err != nil {
			return err
		}
	}

	config.PublicIPRequired = "true"
	config.Count = 1
	config.Hostname = civoProvider.Config.MachineID
	config.Size = civoProvider.Config.MachineType
	config.Region = civoProvider.Config.Region
	config.PublicIPRequired = "true"
	userData, err := GetInjectKeypairScript(civoProvider, volume)
	if err != nil {
		return err
	}

	config.Script = userData

	instance, err := civoProvider.Client.CreateInstance(config)
	if err != nil {
		return err
	}

	_, err = civoProvider.Client.AttachVolume(volume.ID, instance.ID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteOrStop(civoProvider *CivoProvider, stop bool) error {
	instance, err := GetDevpodInstance(civoProvider)
	if err != nil {
		return err
	}

	_, err = civoProvider.Client.DeleteInstance(instance.ID)
	if err != nil {
		return err
	}

	// if we're NOT stopping, then we're deleting, let's wipe volumes.
	if !stop {
		volume, err := GetVolume(civoProvider)
		if err != nil {
			return err
		}

		_, err = civoProvider.Client.DeleteVolume(volume.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func Status(civoProvider *CivoProvider) (client.Status, error) {
	instance, err := GetDevpodInstance(civoProvider)
	if err != nil {
		// If the instance is removed but volume is there
		// it means it is only stopped
		_, err2 := GetVolume(civoProvider)
		if err2 != nil {
			return client.StatusNotFound, nil
		}

		return client.StatusStopped, nil
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
