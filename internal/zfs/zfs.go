// Package zfs adapts the zpool and zfs CLI tools into typed operations. It is
// a leaf package: it shells out to the host's ZFS tooling and has no knowledge
// of application config or orchestration.
package zfs

import (
	"fmt"
	"strings"

	"pb2p/internal/exec"
)

// EnsureLogicalVolume creates the LVM logical volume on the "pve" volume group
// if it does not already exist. It is idempotent.
func EnsureLogicalVolume(lvSize, lvName string) error {
	exists, err := logicalVolumeExists(lvName)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Logical volume %s already exists, skipping creation...\n", lvName)
		return nil
	}

	if _, err := exec.Run("lvcreate", "--yes", "--wipesignatures", "y",
		"-L", lvSize, "-n", lvName, "pve"); err != nil {
		return fmt.Errorf("creating logical volume %s: %w", lvName, err)
	}
	return nil
}

func EnsurePool(poolName, mountPoint, devicePath string) error {
	exists, err := poolExists(poolName)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Pool %s already exists, skipping creation...\n", poolName)
		return nil
	}

	if _, err := exec.Run("zpool", "create", poolName, "-m", mountPoint, devicePath); err != nil {
		return fmt.Errorf("creating zpool %s: %w", poolName, err)
	}
	return nil
}

// EnsureDataset creates the ZFS dataset "poolName/datasetName" if it does not
// already exist. It is idempotent.
func EnsureDataset(poolName, datasetName string) error {
	datasetPath := fmt.Sprintf("%s/%s", poolName, datasetName)
	exists, err := datasetExists(datasetPath)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Dataset %s already exists, skipping creation...\n", datasetPath)
		return nil
	}

	if _, err := exec.Run("zfs", "create", datasetPath); err != nil {
		return fmt.Errorf("creating dataset %s: %w", datasetPath, err)
	}
	return nil
}

// RemoveStaleChunks deletes the PBS chunk store data (and its bookkeeping
// files) that a previous PBS install may have left behind in a dataset. It is
// intended to be run only when the caller wants to reuse an existing pool and
// is okay with discarding its prior contents.
func RemoveStaleChunks(poolName, datasetName string) error {
	datasetPath := fmt.Sprintf("/mnt/%s/%s", poolName, datasetName)

	if !confirmDelete(datasetPath) {
		return fmt.Errorf("aborted: refusing to delete contents of %s", datasetPath)
	}

	fmt.Printf("Removing stale PBS chunks from %s...\n", datasetPath)

	if _, err := exec.Run("rm", "-rf",
		datasetPath+"/.chunks",
		datasetPath+"/.gc-status",
		datasetPath+"/.lock",
	); err != nil {
		return fmt.Errorf("removing stale chunks from %s: %w", datasetPath, err)
	}
	return nil
}

// confirmDelete prompts the user on stdin and returns true only for an
// explicit "y"/"yes" response.
func confirmDelete(datasetPath string) bool {
	fmt.Printf("This will delete the contents of dataset %s, including any backed-up data.\n", datasetPath)
	fmt.Print("Are you sure you want to continue? [y/N]: ")

	var answer string
	fmt.Scanln(&answer)
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

// EnsureProperty sets a single ZFS property (e.g. "compression=lz4") on the
// given dataset only if it isn't already set to that value. It is idempotent.
func EnsureProperty(poolName, datasetName, property string) error {
	datasetPath := fmt.Sprintf("%s/%s", poolName, datasetName)

	// property is of the form "key=value"; split for the existence check.
	key, value, ok := strings.Cut(property, "=")
	if !ok {
		return fmt.Errorf("invalid property %q (expected key=value)", property)
	}

	current, err := getProperty(datasetPath, key)
	if err != nil {
		return err
	}
	if current == value {
		fmt.Printf("Property %s on %s already set, skipping...\n", key, datasetPath)
		return nil
	}

	if _, err := exec.Run("zfs", "set", property, datasetPath); err != nil {
		return fmt.Errorf("setting %s on %s: %w", property, datasetPath, err)
	}
	return nil
}

func poolExists(poolName string) (bool, error) {
	out, err := exec.Run("zpool", "list", "-H", "-o", "name", poolName)
	if err != nil {
		// `zpool list <missing>` exits non-zero; a real (transient) failure is
		// indistinguishable from "not found" at this layer, so treat any error
		// as "does not exist".
		return false, nil
	}
	return strings.TrimSpace(out) == poolName, nil
}

func datasetExists(datasetPath string) (bool, error) {
	out, err := exec.Run("zfs", "list", "-H", "-o", "name", datasetPath)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(out) == datasetPath, nil
}

func getProperty(datasetPath, key string) (string, error) {
	out, err := exec.Run("zfs", "get", "-H", "-o", "value", key, datasetPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func logicalVolumeExists(lvName string) (bool, error) {
	out, err := exec.Run("lvs", "pve", "-o", "lv_name", "--noheadings")
	if err != nil {
		return false, err
	}
	for line := range strings.SplitSeq(out, "\n") {
		if strings.TrimSpace(line) == lvName {
			return true, nil
		}
	}
	return false, nil
}
