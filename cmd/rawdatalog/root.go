package rawdatalog

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "raw-data-log",
	Short: "Raw data log",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(serverCMD)
	RootCmd.AddCommand(readLogsCMD)
}
