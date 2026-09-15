package pve

import (
	"fmt"
	"pbs-setup/internal/exec"
	"strconv"
	"strings"
)

// EnsureContainer creates the container described by spec if it doesn't already
// exist. It is idempotent: an existing container is detected and skipped rather
// than treated as an error.
func EnsureContainer(spec ContainerCreateSpec) error {
	exists, err := containerExists(spec.CTID)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Container %d already exists, skipping creation...\n", spec.CTID)
		return nil
	}
	return createContainer(spec)
}

// containerExists reports whether a container with the given CTID exists.
func containerExists(ctid int) (bool, error) {
	ctidStr := strconv.Itoa(ctid)
	out, err := exec.Run("pct", "list")
	if err != nil {
		return false, err
	}
	return containerListed(out, ctidStr), nil
}

// ContainerCreateSpec is the set of values pve needs to create a container. It
// is intentionally decoupled from config so that pve remains a leaf package
// with no knowledge of application configuration.
type ContainerCreateSpec struct {
	CTID             int
	Hostname         string
	Template         string
	ContainerStorage string
	TemplateStorage  string
	RootDiskGB       int
	Cores            int
	MemoryMB         int
	IP               string
	Gateway          string
	Password         string
	Unprivileged     int
	Features         string
}

func createContainer(spec ContainerCreateSpec) error {
	if err := DownloadTemplateIfNotExists(spec.TemplateStorage, spec.Template); err != nil {
		return fmt.Errorf("ensuring template: %w", err)
	}

	templatePath := fmt.Sprintf("%s:vztmpl/%s", spec.TemplateStorage, spec.Template)
	_, err := exec.Run("pct", "create", strconv.Itoa(spec.CTID), templatePath,
		"--hostname", spec.Hostname,
		"--storage", spec.ContainerStorage,
		"--rootfs", fmt.Sprintf("%s:%d", spec.ContainerStorage, spec.RootDiskGB),
		"--cores", strconv.Itoa(spec.Cores),
		"--memory", strconv.Itoa(spec.MemoryMB),
		"--net0", fmt.Sprintf("name=eth0,bridge=vmbr0,ip=%s,gw=%s", spec.IP, spec.Gateway),
		"--password", spec.Password,
		"--unprivileged", strconv.Itoa(spec.Unprivileged),
		"--features", spec.Features,
	)

	return err
}

func StartContainer(ctid int) error {
	running, err := containerRunning(ctid)
	if err != nil {
		return err
	}
	if running {
		fmt.Printf("Container %d already running, skipping start...\n", ctid)
		return nil
	}

	_, err = exec.Run("pct", "start", strconv.Itoa(ctid))
	return err
}

// containerRunning reports whether the container with the given CTID is in the
// "running" state, according to `pct list`.
func containerRunning(ctid int) (bool, error) {
	ctidStr := strconv.Itoa(ctid)
	out, err := exec.Run("pct", "list")
	if err != nil {
		return false, err
	}
	return containerStatus(out, ctidStr, "running"), nil
}

// containerStatus reports whether pctListOutput contains a row whose first
// field is ctid and whose second field is status.
func containerStatus(pctListOutput, ctid, status string) bool {
	for line := range strings.SplitSeq(pctListOutput, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == ctid && fields[1] == status {
			return true
		}
	}
	return false
}

func ExecInContainer(ctid int, command string, args ...string) (string, error) {
	argv := append([]string{"exec", strconv.Itoa(ctid), "--", command}, args...)
	return exec.Run("pct", argv...)
}

func containerListed(pctListOutput, ctid string) bool {
	for line := range strings.SplitSeq(pctListOutput, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == ctid {
			return true
		}
	}
	return false
}

func templateExists(storage, templateName string) (bool, error) {
	out, err := exec.Run("pveam", "list", storage)
	if err != nil {
		return false, fmt.Errorf("listing templates on %s: %w", storage, err)
	}
	return strings.Contains(out, templateName), nil
}
