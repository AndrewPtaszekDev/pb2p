package cmd

import (
	"fmt"
	"pbs-setup/internal/config"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Create a default config.yaml",
	Long: `Writes a default config.yaml to the current directory.

Existing config.yaml files are left untouched.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.CreateDefaultConfig(); err != nil {
			return fmt.Errorf("creating default config.yaml: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
