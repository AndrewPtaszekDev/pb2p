package cmd

import (
	"fmt"
	"pb2p/internal/config"
	"pb2p/internal/pbs"
	"pb2p/internal/setup"

	"github.com/spf13/cobra"
)

// pbsCmd represents the pbs command
var pbsCmd = &cobra.Command{
	Use:   "pbs",
	Short: "Install PBS and create a peer config",
	Long: `Installs PBS in the running container, creates the datastore and a backup
user, then writes a peer config file for the peer to consume.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("validating config: %w", err)
		}

		peerCfgSpec, err := setup.PBS(cfg.Container.CTID, cfg)
		if err != nil {
			return err
		}

		peerCfg := config.PeerConfig{
			PBSName:       "peer-pbs",
			TokenSecret:   peerCfgSpec.TokenSecret,
			Fingerprint:   peerCfgSpec.Fingerprint,
			Username:      pbs.GetAuthID(cfg.PBS.PeerUsername),
			DatastoreName: cfg.ZFS.DatasetName,
		}

		peerConfigPath := fmt.Sprintf("%s.yaml", cfg.PBS.PeerUsername)
		if err := config.SetPeerConfig(peerConfigPath, peerCfg); err != nil {
			return fmt.Errorf("writing peer config for %s: %w", cfg.PBS.PeerUsername, err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pbsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pbsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pbsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
