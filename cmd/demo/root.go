package cmd

import (
	"fmt"
	"os"

	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/config"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "integrate-vault-with-microservices-in-k8s",
	Short: "A brief description of your application",
	Long:  `A longer description that spans multiple lines`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnFinalize(initConfig)
}

func initConfig() {
	config.Init()
}
