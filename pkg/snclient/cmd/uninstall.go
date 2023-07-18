//go:build windows

package cmd

import (
	"os"

	"pkg/snclient"

	"github.com/spf13/cobra"
)

func init() {
	uninstallCmd := &cobra.Command{
		Use:   "uninstall [cmd]",
		Short: "Uninstall windows service and firewall exception",
		Long: `Uninstall is used during msi installation for removing the windows service and a firewall exception.
`,
	}
	rootCmd.AddCommand(uninstallCmd)

	// uninstall stop
	uninstallCmd.AddCommand(&cobra.Command{
		Use:   "stop [args]",
		Short: "called from the msi installer stop the service",
		Run: func(cmd *cobra.Command, args []string) {
			agentFlags.Mode = snclient.ModeOneShot
			snc := snclient.NewAgent(agentFlags)

			snc.Log.Infof("uninstaller: stop")
			if hasService(WINSERVICE) {
				err := stopService("snclient")
				if err != nil {
					snc.Log.Errorf("failed to stops service: %s", err.Error())
				}
			}
			snc.Log.Infof("stop completed")

			os.Exit(0)
		},
	})

	// uninstall pkg
	uninstallCmd.AddCommand(&cobra.Command{
		Use:   "pkg [args]",
		Short: "called from the msi installer, removes firewall and service if agent gets uninstalled",
		Run: func(cmd *cobra.Command, args []string) {
			agentFlags.Mode = snclient.ModeOneShot
			snc := snclient.NewAgent(agentFlags)

			installConfig := parseInstallerArgs(args)
			if installConfig["REMOVE"] != "ALL" {
				snc.Log.Infof("skipping uninstall: %#v", installConfig)
				os.Exit(0)
			}

			snc.Log.Infof("starting uninstaller: %#v", installConfig)
			if hasService(WINSERVICE) {
				err := stopService("snclient")
				if err != nil {
					snc.Log.Errorf("failed to stops service: %s", err.Error())
				}
				err = removeService("snclient")
				if err != nil {
					snc.Log.Errorf("failed to remove service: %s", err.Error())
				}
			}

			snc.Log.Infof("uninstall completed")
			os.Exit(0)
		},
	})
}