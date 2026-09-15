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

	// Add encrpytion

	return nil
}
