package cmd

import (
	"fmt"
	"pb2p/internal/config"
	"pb2p/internal/setup"

	"github.com/spf13/cobra"
)

// zfsCmd represents the zfs command
var zfsCmd = &cobra.Command{
	Use:   "zfs",
	Short: "Create the ZFS storage backing for PBS",
	Long: `Provisions an LVM logical volume, a zpool on top of it, and a dataset
with the desired properties.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w\nRun pb2p config first", err)
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("validating config: %w", err)
		}

		if err := setup.ZFS(cfg.ZFS, rmStaleChunks); err != nil {
			return fmt.Errorf("setting up ZFS pool %s: %w", cfg.ZFS.PoolName, err)
		}
		return nil
	},
}

var rmStaleChunks bool

func init() {
	rootCmd.AddCommand(zfsCmd)

	zfsCmd.Flags().BoolVar(&rmStaleChunks, "rm-stale-chunks", false,
		"remove stale PBS chunk data left behind in the dataset by a previous PBS setup. "+
			"Only use this when reusing an existing pool and you are okay with discarding its contents")
}
