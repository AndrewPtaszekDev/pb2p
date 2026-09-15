package cmd

import (
	"fmt"
	"pbs-setup/internal/config"
	"pbs-setup/internal/setup"

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
			return fmt.Errorf("loading config: %w", err)
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("validating config: %w", err)
		}

		if err := setup.ZFS(cfg.ZFS); err != nil {
			return fmt.Errorf("setting up ZFS pool %s: %w", cfg.ZFS.PoolName, err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(zfsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// zfsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// zfsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
