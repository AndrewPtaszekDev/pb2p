package setup

import (
	"fmt"

	"pbs-setup/internal/config"
	"pbs-setup/internal/pbs"
)

// PeerConfigSpec is the subset of a peer config that setup.PBS produces: the
// token secret and the PBS host certificate fingerprint.
type PeerConfigSpec struct {
	TokenSecret string
	Fingerprint string
}

// PBS provisions the PBS server inside an already-running container: it
// installs PBS, creates the datastore and its jobs, then creates a backup user
// (with token + permissions) and returns the values the peer needs.
func PBS(ctid int, cfg config.Config) (PeerConfigSpec, error) {
	if err := pbs.EnsurePBS(ctid); err != nil {
		return PeerConfigSpec{}, fmt.Errorf("installing PBS in container %d: %w", ctid, err)
	}

	if err := pbs.EnsureDatastore(ctid, cfg.ZFS.PoolName, cfg.ZFS.DatasetName); err != nil {
		return PeerConfigSpec{}, fmt.Errorf("adding datastore %s: %w", cfg.ZFS.DatasetName, err)
	}

	tokenSecret, err := createUserWithPerms(ctid, cfg.PBS.PeerUsername, cfg.ZFS.DatasetName)
	if err != nil {
		return PeerConfigSpec{}, fmt.Errorf("creating user/token perms for %s: %w", cfg.PBS.PeerUsername, err)
	}

	fingerprint, err := pbs.GetFingerprint(ctid)
	if err != nil {
		return PeerConfigSpec{}, fmt.Errorf("getting PBS fingerprint from container %d: %w", ctid, err)
	}

	return PeerConfigSpec{
		TokenSecret: tokenSecret,
		Fingerprint: fingerprint,
	}, nil
}

// createUserWithPerms creates the backup user, its token, and its permissions.
// All steps are idempotent.
func createUserWithPerms(ctid int, peerUsername, datasetName string) (string, error) {
	username := peerUsername + "@pbs"

	if err := pbs.EnsureUser(ctid, username); err != nil {
		return "", err
	}

	secret, err := pbs.EnsureToken(ctid, username, pbs.TokenName, func() (string, error) {
		existing, readErr := config.GetYAMLKey(peerUsername+".yaml", "token_secret")
		if readErr != nil {
			return "", readErr
		}
		return existing, nil
	})
	if err != nil {
		return "", err
	}

	authID := pbs.GetAuthID(peerUsername)
	if err := pbs.EnsurePermissions(ctid, username, authID, datasetName); err != nil {
		return "", err
	}

	return secret, nil
}
