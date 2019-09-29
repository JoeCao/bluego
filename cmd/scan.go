package cmd

import (
	"bluego/discovery"
	"github.com/spf13/cobra"
)

// discoveryCmd represents the discovery command
var discoveryCmd = &cobra.Command{
	Use:   "discovery",
	Short: "bluetooth discovery example",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		adapterID, err := cmd.Flags().GetString("adapterID")
		if err != nil {
			fail(err)
		}

		onlyBeacon, err := cmd.Flags().GetBool("beacon")
		if err != nil {
			fail(err)
		}

		fail(discovery.Run(adapterID, onlyBeacon))
	},
}

func init() {
	rootCmd.AddCommand(discoveryCmd)
	discoveryCmd.Flags().BoolP("beacon", "b", false, "Only report beacons")
}
