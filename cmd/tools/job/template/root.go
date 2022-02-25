package template

import "github.com/spf13/cobra"

var RootCMD = &cobra.Command{
	Use:   "template",
	Short: "Commands to create pre-filled templates for jobs",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(customerCMD)
	RootCMD.AddCommand(applicationCMD)
}
