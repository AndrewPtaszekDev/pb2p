package setup

import (
	"fmt"

	"pb2p/internal/config"
	"pb2p/internal/pve"
)

// Peer registers a peer's PBS storage on this host.
func Peer(cfg config.PeerConfig) error {
	if err := pve.EnsurePBSStorage(
		cfg.PBSName,
		cfg.IP,
		cfg.Username,
		cfg.TokenSecret,
		cfg.DatastoreName,
		cfg.Fingerprint,
	); err != nil {
		return fmt.Errorf("adding PBS storage: %w", err)
	}

	// Put the encryption key in a file at the call site
	keyPath, err := pve.ExportEncryptionKey(cfg.PBSName)
	if err != nil {
		return fmt.Errorf("exporting encryption key: %w", err)
	}

	fmt.Printf("Encryption key saved to %s\n", keyPath)
	fmt.Println("Back it up somewhere other than this host (password manager, USB drive, printed copy).")
	fmt.Println("Backups on your peer are encrypted with it; if this host is lost, they cannot be restored without it.")

	return nil
}
