package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the full server setup",
	Long: `Runs each setup step in order:

	config --> zfs --> container --> pbs

Each step runs exactly as if it were invoked directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		subcommands := []string{"zfs", "container", "pbs"}
		for _, name := range subcommands {
			fmt.Printf("=== running %s ===\n", name)
			if err := runSubcommand(cmd, name); err != nil {
				return err
			}
		}

		return nil
	},
}

func runSubcommand(cmd *cobra.Command, name string) error {
	cmd.Root().SetArgs([]string{name})
	return cmd.Root().Execute()
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
