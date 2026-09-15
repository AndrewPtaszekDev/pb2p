package pbs

import (
	"fmt"
	"pbs-setup/internal/pve"
)

const TokenName = "backup-access"

func EnsurePBS(ctid int) error {
	if pbsExists(ctid) {
		fmt.Printf("PBS on CTID: %d already installed, skipping...\n", ctid)
		return nil
	}

	if _, err := pve.ExecInContainer(ctid, "bash", "-c",
		`echo "deb http://download.proxmox.com/debian/pbs bookworm pbs-no-subscription" > /etc/apt/sources.list.d/pbs.list`,
	); err != nil {
		return fmt.Errorf("add pbs repo: %w", err)
	}

	if _, err := pve.ExecInContainer(ctid, "wget",
		"https://enterprise.proxmox.com/debian/proxmox-release-bookworm.gpg",
		"-O", "/etc/apt/trusted.gpg.d/proxmox-release-bookworm.gpg",
	); err != nil {
		return fmt.Errorf("add signing key: %w", err)
	}

	if _, err := pve.ExecInContainer(ctid, "apt", "update"); err != nil {
		return fmt.Errorf("apt update: %w", err)
	}

	if _, err := pve.ExecInContainer(ctid, "apt", "install", "-y", "proxmox-backup-server"); err != nil {
		return fmt.Errorf("install pbs: %w", err)
	}

	return nil
}

// TODO: technically this is dropping a potential error on the command (the DOES exist but errors case)
func pbsExists(ctid int) bool {
	_, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "version")
	if err == nil {
		return true
	}
	return false
}
