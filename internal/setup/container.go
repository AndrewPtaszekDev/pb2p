package setup

import (
	"fmt"

	"pb2p/internal/config"
	"pb2p/internal/pve"
)

// Container provisions a container from cfg: it ensures the container exists
// (creating it, along with its template, if needed), attaches the ZFS bind
// mount, and starts it.
func Container(cfg config.ContainerConfig, zpoolName, datasetName string) error {
	spec := pve.ContainerCreateSpec{
		CTID:             cfg.CTID,
		Hostname:         cfg.Hostname,
		Template:         cfg.Template,
		ContainerStorage: cfg.ContainerStorage,
		TemplateStorage:  cfg.TemplateStorage,
		RootDiskGB:       cfg.RootDiskGB,
		Cores:            cfg.Cores,
		MemoryMB:         cfg.MemoryMB,
		IP:               cfg.IP,
		Gateway:          cfg.Gateway,
		Password:         cfg.Password,
		Unprivileged:     0,
		Features:         "nesting=1",
	}

	if err := pve.EnsureContainer(spec); err != nil {
		return fmt.Errorf("ensuring container %d: %w", cfg.CTID, err)
	}

	if err := pve.SetBindMount(cfg.CTID, zpoolName, datasetName); err != nil {
		return fmt.Errorf("setting bind mount for container %d: %w", cfg.CTID, err)
	}

	if err := pve.StartContainer(cfg.CTID); err != nil {
		return fmt.Errorf("starting container %d: %w", cfg.CTID, err)
	}

	return nil
}
