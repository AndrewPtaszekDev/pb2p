package cmd

import (
	"fmt"
	"pb2p/internal/config"
	"pb2p/internal/setup"

	"github.com/spf13/cobra"
)

var peerCmd = &cobra.Command{
	Use:   "peer <peer-config.yaml>",
	Short: "Register a peer's PBS storage",
	Long: `Reads a peer config file and adds the peer's PBS storage on this host.

Pass the path to the peer config you received from your peer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expected a single argument: peerConfigPath")
		}
		peerConfigPath := args[0]

		cfg, err := config.GetPeerConfig(peerConfigPath)
		if err != nil {
			return fmt.Errorf("loading peer config for %s: %w", peerConfigPath, err)
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("validating peer config: %w", err)
		}

		if err := setup.Peer(cfg); err != nil {
			return fmt.Errorf("adding PBS storage for peer %s: %w", cfg.IP, err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(peerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
