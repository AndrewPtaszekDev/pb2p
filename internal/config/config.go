package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"

	"go.yaml.in/yaml/v3"
	//"github.com/stretchr/testify/assert/yaml"
)

//go:embed defaults.yaml
var defaultConfigData []byte

const configFileName = "config.yaml"

type Config struct {
	Container ContainerConfig `yaml:"container"`
	ZFS       ZFSConfig       `yaml:"zfs"`
	PBS       PBSConfig       `yaml:"pbs"`
}

type ContainerConfig struct {
	CTID             int    `yaml:"ctid"`
	Hostname         string `yaml:"hostname"`
	Template         string `yaml:"template"`
	ContainerStorage string `yaml:"container_storage"`
	TemplateStorage  string `yaml:"template_storage"`
	RootDiskGB       int    `yaml:"root_disk_gb"`
	Cores            int    `yaml:"cores"`
	MemoryMB         int    `yaml:"memory_mb"`
	IP               string `yaml:"ip,omitempty"`
	Gateway          string `yaml:"gateway,omitempty"`
	Password         string `yaml:"password"`
}

type ZFSConfig struct {
	LVsize      string `yaml:"lvsize"`
	LVName      string `yaml:"lvname"`
	PoolName    string `yaml:"pool_name"`
	DatasetName string `yaml:"dataset_name"`
}

type PBSConfig struct {
	PeerUsername string `yaml:"peer_username"`
}

// what the peer needs to give to its pb2p client --config /path/config.yaml
type PeerConfig struct {
	PBSName       string `yaml:"pbs_name"` // peer-pbs
	IP            string `yaml:"ip"`
	TokenSecret   string `yaml:"token_secret"`
	Fingerprint   string `yaml:"fingerprint"`
	Username      string `yaml:"username"`
	DatastoreName string `yaml:"datastore_name"`
}

func (c Config) Validate() error {
	if c.Container.CTID == 0 {
		return fmt.Errorf("container.ctid is required")
	}
	if c.Container.Password == "" {
		return fmt.Errorf("container.password is required")
	}
	if c.Container.IP == "" {
		return fmt.Errorf("container.ip is required")
	}
	if c.Container.Gateway == "" {
		return fmt.Errorf("container.gateway is required")
	}
	if c.PBS.PeerUsername == "" {
		return fmt.Errorf("pbs.peer_username is required")
	}
	if c.ZFS.DatasetName == "" {
		return fmt.Errorf("zfs.dataset_name is required")
	}
	return nil
}

func (c PeerConfig) Validate() error {
	if c.IP == "" {
		return fmt.Errorf("ip is required. ip is the location of your peer's PBS instance (probably via WG)")
	}
	return nil
}

func CreateDefaultConfig() error {
	// O_CREATE|O_EXCL makes the existence check and the write a single atomic
	// operation, so we never clobber a config that appeared in between.
	err := writeFileAtomic(configFileName, defaultConfigData, 0644, true)
	if errors.Is(err, fs.ErrExist) {
		fmt.Printf("%s already exists, skipping...\n", configFileName)
		return nil
	}

	return err
}

// GetConfig loads config.yaml.
func GetConfig() (Config, error) {
	data, err := os.ReadFile(configFileName)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// GetPeerConfig loads the peer config written by SetPeerConfig, which is named
// after the peer's username.
func GetPeerConfig(peerConfigPath string) (PeerConfig, error) {
	data, err := os.ReadFile(peerConfigPath)
	if err != nil {
		return PeerConfig{}, err
	}

	var cfg PeerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return PeerConfig{}, err
	}

	return cfg, nil
}

// SetPeerConfig writes the peer config that the peer feeds back into its
// config.yaml. The file is created if it does not exist yet.
//
// ip is intentionally left empty: it is filled in by the peer.
func SetPeerConfig(peerConfigPath string, cfg PeerConfig) error {
	data, err := readYAMLMap(peerConfigPath)
	if err != nil {
		return err
	}

	// ip is intentionally left empty: it is filled in by the peer.
	maps.Copy(data, map[string]any{
		"pbs_name":       cfg.PBSName,
		"ip":             "",
		"username":       cfg.Username,
		"token_secret":   cfg.TokenSecret,
		"fingerprint":    cfg.Fingerprint,
		"datastore_name": cfg.DatastoreName,
	})

	return writeYAMLMap(peerConfigPath, data)
}
