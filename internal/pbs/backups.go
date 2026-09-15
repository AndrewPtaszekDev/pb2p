package pbs

// pbs adapts the proxmox-backup-manager CLI, running commands inside a
// container via pve.ExecInContainer.

import (
	"encoding/json"
	"fmt"
	"pb2p/internal/exec"
	"pb2p/internal/pve"
	"strings"
)

func EnsureDatastore(ctid int, poolName string, datasetName string) error {
	backingPath := fmt.Sprintf("/mnt/%s/%s", poolName, datasetName)
	if err := ensureDatastore(ctid, datasetName, backingPath); err != nil {
		return err
	}

	pruneJobID := fmt.Sprint(datasetName, "-prune")
	if err := ensurePruneJob(ctid, pruneJobID, datasetName); err != nil {
		return err
	}

	verifyJobID := fmt.Sprint(datasetName, "-verify")
	if err := ensureVerifyJob(ctid, verifyJobID, datasetName); err != nil {
		return err
	}

	return nil
}

func EnsureUser(ctid int, username string) error {
	exists, err := userExists(ctid, username)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("User %s already exists, skipping creation...\n", username)
		return nil
	}

	if _, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "user", "create", username); err != nil {
		return err
	}
	return nil
}

func EnsureToken(ctid int, username, tokenName string, recoverSecret func() (string, error)) (string, error) {
	secret, err := GenerateToken(ctid, username, tokenName)
	if err == nil {
		return secret, nil
	}
	if !exec.IsAlreadyExists(err) {
		return "", err
	}

	// Token already exists; recover its secret if it was persisted.
	if recoverSecret != nil {
		if existing, readErr := recoverSecret(); readErr == nil && existing != "" {
			return existing, nil
		}
	}

	// Secret lost; delete and regenerate.
	if err := deleteToken(ctid, username, tokenName); err != nil {
		return "", fmt.Errorf("token exists but secret is lost, and delete-token failed: %w", err)
	}
	secret, err = GenerateToken(ctid, username, tokenName)
	if err != nil {
		return "", fmt.Errorf("failed to regenerate token after delete: %w", err)
	}
	return secret, nil
}

// GenerateToken creates an API token for a user and returns its secret.
func GenerateToken(ctid int, username, tokenName string) (string, error) {
	tokenOutput, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "user", "generate-token",
		username,
		tokenName,
	)
	if err != nil {
		return "", err
	}
	return parseTokenOutput(tokenOutput)
}

func deleteToken(ctid int, username, tokenName string) error {
	_, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "user", "delete-token", username, tokenName)
	return err
}

func EnsurePermissions(ctid int, username, authID, datasetName string) error {
	backingPath := fmt.Sprintf("/datastore/%s", datasetName)
	if err := ensureACL(ctid, username, "DatastoreBackup", backingPath); err != nil {
		return err
	}
	if err := ensureACL(ctid, authID, "DatastoreBackup", backingPath); err != nil {
		return err
	}
	return nil
}

func GetFingerprint(ctid int) (string, error) {
	out, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "cert", "info")
	if err != nil {
		return "", fmt.Errorf("reading PBS certificate info: %w", err)
	}

	const prefix = "Fingerprint (sha256):"
	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}

		fingerprint := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if fingerprint == "" {
			return "", fmt.Errorf("empty fingerprint in cert output: %s", out)
		}

		return fingerprint, nil
	}

	return "", fmt.Errorf("fingerprint not found in cert output: %s", out)
}

func GetAuthID(peerUsername string) string {
	return fmt.Sprint(peerUsername, "@pbs!", TokenName)
}

func ensureDatastore(ctid int, datasetName, backingPath string) error {
	exists, err := datastoreExists(ctid, datasetName)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Datastore %s already exists, skipping creation...\n", datasetName)
		return nil
	}

	if _, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "datastore", "create",
		datasetName,
		backingPath,
		"--gc-schedule", "daily",
	); err != nil {
		if strings.Contains(err.Error(), "EEXIST") || strings.Contains(err.Error(), ".chunks") {
			return fmt.Errorf("%w\nThe backing path already contains data from a previous PBS setup. "+
				"If you want to reuse this dataset and are okay with discarding its contents, "+
				"re-run with: pb2p zfs --rm-stale-chunks", err)
		}
		return err
	}
	return nil
}

func ensurePruneJob(ctid int, jobID, datasetName string) error {
	if ok, err := pruneJobExists(ctid, jobID); err != nil {
		return err
	} else if ok {
		fmt.Printf("Prune job %s already exists, skipping creation...\n", jobID)
		return nil
	}

	if _, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "prune-job", "create",
		jobID,
		"--store", datasetName,
		"--schedule", "daily",
		"--keep-last", "7",
		"--keep-weekly", "4",
		"--keep-monthly", "3",
	); err != nil {
		return err
	}
	return nil
}

func ensureVerifyJob(ctid int, jobID, datasetName string) error {
	if ok, err := verifyJobExists(ctid, jobID); err != nil {
		return err
	} else if ok {
		fmt.Printf("Verify job %s already exists, skipping creation...\n", jobID)
		return nil
	}

	if _, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "verify-job", "create",
		jobID,
		"--store", datasetName,
		"--schedule", "daily",
		"--outdated-after", "30",
	); err != nil {
		return err
	}
	return nil
}

func ensureACL(ctid int, authID, role, backingPath string) error {
	exists, err := aclExists(ctid, authID, role, backingPath)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("ACL for %s already set, skipping...\n", authID)
		return nil
	}

	if _, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "acl", "update",
		backingPath,
		role,
		"--auth-id", authID,
	); err != nil {
		return err
	}
	return nil
}

func datastoreExists(ctid int, datasetName string) (bool, error) {
	out, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "datastore", "list")
	if err != nil {
		return false, err
	}
	return strings.Contains(out, datasetName), nil
}

func pruneJobExists(ctid int, jobID string) (bool, error) {
	out, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "prune-job", "list")
	if err != nil {
		return false, err
	}
	return strings.Contains(out, jobID), nil
}

func verifyJobExists(ctid int, jobID string) (bool, error) {
	out, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "verify-job", "list")
	if err != nil {
		return false, err
	}
	return strings.Contains(out, jobID), nil
}

func userExists(ctid int, username string) (bool, error) {
	out, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "user", "list")
	if err != nil {
		return false, err
	}
	return strings.Contains(out, username), nil
}

func aclExists(ctid int, authID, role, backingPath string) (bool, error) {
	out, err := pve.ExecInContainer(ctid, "proxmox-backup-manager", "acl", "list")
	if err != nil {
		return false, err
	}
	// A matching ACL row starts with the path and lists the auth-id and role in
	// later fields. Match against the whole row: the path and role must both
	// appear, and the auth-id must be present.
	return aclRowMatches(out, authID, role, backingPath), nil
}

// aclRowMatches reports whether any ACL list row matches the given path, role,
// and auth-id. The row format is "<path> <role> <auth-id>", but the order of
// role/auth-id is not guaranteed, so we match on presence rather than column.
func aclRowMatches(out, authID, role, backingPath string) bool {
	for line := range strings.SplitSeq(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != backingPath {
			continue
		}
		hasRole := false
		hasAuthID := false
		for _, f := range fields[1:] {
			if f == role {
				hasRole = true
			}
			if f == authID {
				hasAuthID = true
			}
		}
		if hasRole && hasAuthID {
			return true
		}
	}
	return false
}

type tokenCreateResult struct {
	TokenID string `json:"tokenid"`
	Value   string `json:"value"`
}

func parseTokenOutput(output string) (string, error) {
	idx := strings.Index(output, "{")
	if idx == -1 {
		return "", fmt.Errorf("no JSON object found in token output: %s", output)
	}
	trimmed := output[idx:]

	var result tokenCreateResult
	if err := json.Unmarshal([]byte(trimmed), &result); err != nil {
		return "", fmt.Errorf("failed to parse token output: %w\nraw output: %s", err, output)
	}
	if result.Value == "" {
		return "", fmt.Errorf("token value missing from output: %s", output)
	}
	return result.Value, nil
}
