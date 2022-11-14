package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/undeadops/fh/pkg/contract"
	"github.com/undeadops/fh/pkg/destroy"
	"github.com/undeadops/fh/pkg/get"
	"github.com/undeadops/fh/pkg/up"
)

var (
	org        string
	debug      bool
	region     string
	configFile string
)

func configureCLI() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:  "fh",
		Long: "Deploy your applications",
	}

	rootCommand.AddCommand(up.Command())
	rootCommand.AddCommand(destroy.Command())
	rootCommand.AddCommand(get.Command())

	rootCommand.PersistentFlags().StringVarP(&configFile, "values", "f", "", "Deployment values file")
	rootCommand.PersistentFlags().StringVarP(&org, "org", "o", "", "Pulumi org to use for your stack")
	rootCommand.PersistentFlags().StringVarP(&region, "region", "r", "us-east-2", "AWS Region to use")
	rootCommand.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")

	viper.BindEnv("region", "AWS_REGION") // if the user has set the AWS_REGION env var, use it

	viper.BindPFlag("org", rootCommand.PersistentFlags().Lookup("org"))
	viper.BindPFlag("region", rootCommand.PersistentFlags().Lookup("region"))
	viper.BindPFlag("values", rootCommand.PersistentFlags().Lookup("values"))

	return rootCommand
}

func init() {
	log.SetLevel(log.InfoLevel)
	cobra.OnInitialize(initConfig)
}

func initConfig() {

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.fh")
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file: ", viper.ConfigFileUsed())
	}
}

func main() {
	rootCommand := configureCLI()

	if err := rootCommand.Execute(); err != nil {
		contract.IgnoreIoError(fmt.Fprintf(os.Stderr, "%s", err))
		os.Exit(1)
	}
}
