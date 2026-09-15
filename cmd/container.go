package cmd

import (
	"fmt"
	"pbs-setup/internal/config"
	"pbs-setup/internal/setup"

	"github.com/spf13/cobra"
)

// containerCmd represents the container command
var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Create and start the PBS container",
	Long: `Creates the container (and its template, if needed), attaches the ZFS
bind mount, and starts it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("validating config: %w", err)
		}

		if err := setup.Container(cfg.Container, cfg.ZFS.PoolName, cfg.ZFS.DatasetName); err != nil {
			return fmt.Errorf("creating container: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(containerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// containerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// containerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
