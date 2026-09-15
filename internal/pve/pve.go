package pve

import (
	"fmt"
	"os"
	"pb2p/internal/exec"
	"strings"
)

func EnsurePBSStorage(pbsName, ip, username, tokenSecret, datastoreName, fingerprint string) error {
	_, err := exec.Run("pvesm", "add", "pbs", pbsName,
		"--server", ip,
		"--username", username,
		"--password", tokenSecret,
		"--datastore", datastoreName,
		"--fingerprint", fingerprint,
		"--encryption-key", "autogen",
	)

	if exec.IsAlreadyExists(err) {
		fmt.Println("PBS storage already exists, skipping...")
		return nil
	}

	return err
}

func DownloadTemplateIfNotExists(templateStorage, template string) error {
	exists, err := templateExists(templateStorage, template)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// Refresh the appliance index so pveam knows about current templates
	if _, err := exec.Run("pveam", "update"); err != nil {
		return fmt.Errorf("updating pveam index: %w", err)
	}

	if _, err := exec.Run("pveam", "download", templateStorage, template); err != nil {
		return fmt.Errorf("downloading template %s: %w", template, err)
	}

	return nil
}

func SetBindMount(ctid int, poolName string, datasetName string) error {
	datasetPath := fmt.Sprintf("/mnt/%s/%s", poolName, datasetName)

	lxcConfPath := fmt.Sprintf("/etc/pve/lxc/%d.conf", ctid)

	// Skip if a mount for this dataset is already configured.
	if hasBindMount(lxcConfPath, datasetPath) {
		fmt.Printf("Bind mount for %s already configured, skipping...\n", datasetPath)
		return nil
	}

	mount := fmt.Sprintf("mp0: %s,mp=%s\n", datasetPath, datasetPath)

	f, err := os.OpenFile(lxcConfPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening %s: %w", lxcConfPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(mount); err != nil {
		return fmt.Errorf("writing mp0 line to %s: %w", lxcConfPath, err)
	}

	return nil
}

func hasBindMount(lxcConfPath, datasetPath string) bool {
	data, err := os.ReadFile(lxcConfPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), fmt.Sprintf("mp=%s", datasetPath))
}
