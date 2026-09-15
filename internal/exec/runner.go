package exec

import (
	"fmt"
	"os/exec"
	"strings"
)

func Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s failed: %w\n%s", name, err, string(output))
	}
	return string(output), nil
}

// IsAlreadyExists reports whether err (as produced by Run) indicates the
// underlying operation failed because its target already exists. Run embeds the
// command's combined output in err, so this checks the captured output for the
// "already exists" signal.
func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "already exists")
}
