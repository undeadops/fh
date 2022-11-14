package up

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/events"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	log "github.com/sirupsen/logrus"
	"github.com/undeadops/fh/pkg/pulumi"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/undeadops/fh/pkg/config"
)

var (
	dryrun    bool
	name      string
	directory string
	verbose   bool
	nlb       bool
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "up",
		Short: "Deploy your application",
		Long:  "Deploy your application to Kubernetes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			// Set some required params
			ctx := context.Background()

			config, err := config.NewConfig(viper.GetString("values"))
			if err != nil {
				return fmt.Errorf("error parsing config: %v", err)
			}
			org := config.Org
			region := config.Aws.Region
			name := config.Slug

			fmt.Printf("Region: %s - Name: %s", region, name)
			stackName := auto.FullyQualifiedStackName(org, "fh", name)

			pulumiStack, err := auto.UpsertStackInlineSource(ctx, stackName, "fh", nil)
			if err != nil {
				return fmt.Errorf("failed to create or select stack: %v", err)
			}

			err = pulumiStack.SetConfig(ctx, "aws:skipMetadataApiCheck", auto.ConfigValue{Value: "false"})
			if err != nil {
				return err
			}

			workspace := pulumiStack.Workspace()

			err = pulumi.EnsurePlugins(workspace)
			if err != nil {
				return err
			}

			workspace.SetProgram(pulumi.Deploy(name, directory, nlb))

			if dryrun {
				_, err = pulumiStack.Preview(ctx, optpreview.Message("Running fh dry-run"))
				if err != nil {
					return fmt.Errorf("error creating stack: %v", err)
				}
			} else {
				var streamer optup.Option
				if verbose {
					streamer = optup.ProgressStreams(os.Stdout)
				} else {
					upChannel := make(chan events.EngineEvent)
					go collectEvents(upChannel)

					streamer = optup.EventStreams(upChannel)
				}

				log.Infof("Creating fh application: %s", name)
				_, err = pulumiStack.Up(ctx, streamer)

				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	f := command.Flags()
	f.BoolVarP(&dryrun, "preview", "p", false, "Preview changes, dry-run mode")
	f.BoolVarP(&verbose, "verbose", "v", false, "Show output of Pulumi operations")
	f.StringVarP(&directory, "dir", "d", ".", "Path to docker context to use")
	f.BoolVar(&nlb, "nlb", false, "Provision an NLB instead of ELB")

	return command
}

func collectEvents(eventChannel <-chan events.EngineEvent) {

	for {

		var event events.EngineEvent
		var ok bool

		createLogger := log.WithFields(log.Fields{"event": "CREATING"})
		completeLogger := log.WithFields(log.Fields{"event": "COMPLETE"})

		event, ok = <-eventChannel
		if !ok {
			return
		}

		if event.ResourcePreEvent != nil {

			switch event.ResourcePreEvent.Metadata.Type {
			case "aws:ecr/repository:Repository":
				createLogger.WithFields(log.Fields{"resource": event.ResourcePreEvent.Metadata.Type}).Info("Creating ECR repository")
			case "kubernetes:core/v1:Namespace":
				createLogger.WithFields(log.Fields{"resource": event.ResourcePreEvent.Metadata.Type}).Info("Creating Kubernetes Namespace")
			case "kubernetes:core/v1:Service":
				createLogger.WithFields(log.Fields{"resource": event.ResourcePreEvent.Metadata.Type}).Info("Creating Kubernetes Service")
			case "kubernetes:apps/v1:Deployment":
				createLogger.WithFields(log.Fields{"resource": event.ResourcePreEvent.Metadata.Type}).Info("Creating Kubernetes Deployment")
			}
		}

		if event.ResOutputsEvent != nil {
			switch event.ResOutputsEvent.Metadata.Type {
			case "aws:ecr/repository:Repository":
				completeLogger.WithFields(log.Fields{"name": event.ResOutputsEvent.Metadata.New.Outputs["repositoryUrl"], "resource": event.ResOutputsEvent.Metadata.Type}).Info("Created ECR repository")
			case "kubernetes:core/v1:Namespace":
				completeLogger.WithFields(log.Fields{"resource": event.ResOutputsEvent.Metadata.Type}).Info("Created Kubernetes Namespace")
			case "kubernetes:core/v1:Service":
				completeLogger.WithFields(log.Fields{"resource": event.ResOutputsEvent.Metadata.Type}).Info("Created Kubernetes Service")
			case "kubernetes:apps/v1:Deployment":
				completeLogger.WithFields(log.Fields{"resource": event.ResOutputsEvent.Metadata.Type}).Info("Created Kubernetes Deployment")
			}

		}
	}
}
