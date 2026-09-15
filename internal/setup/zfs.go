package setup

import (
	"fmt"

	"pb2p/internal/config"
	"pb2p/internal/zfs"
)

// ZFS provisions the ZFS storage backing for PBS: an LVM logical volume on the
// "pve" volume group, a zpool on top of it, and a dataset with the desired
// properties.
func ZFS(cfg config.ZFSConfig) error {
	if err := zfs.EnsureLogicalVolume(cfg.LVsize, cfg.LVName); err != nil {
		return fmt.Errorf("ensuring logical volume %s: %w", cfg.LVName, err)
	}

	devicePath := fmt.Sprintf("/dev/pve/%s", cfg.LVName)
	mountPoint := fmt.Sprintf("/mnt/%s", cfg.PoolName)

	if err := zfs.EnsurePool(cfg.PoolName, mountPoint, devicePath); err != nil {
		return err
	}

	if err := zfs.EnsureDataset(cfg.PoolName, cfg.DatasetName); err != nil {
		return err
	}

	if err := zfs.EnsureProperty(cfg.PoolName, cfg.DatasetName, "compression=lz4"); err != nil {
		return err
	}

	if err := zfs.EnsureProperty(cfg.PoolName, cfg.DatasetName, "atime=off"); err != nil {
		return err
	}

	return nil
}
