package get

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/prometheus/common/log"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/undeadops/fh/pkg/config"
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "get",
		Short: "Get all ploy deployed applications",
		Long:  "Get all ploy deployed applications",
		RunE: func(cmd *cobra.Command, args []string) error {

			// Parse Values config
			valuesFile := viper.GetString("values")
			config, err := config.NewConfig(valuesFile)
			if err != nil {
				return fmt.Errorf("problem parsing values file: %w", err)
			}

			// Required params
			ctx := context.Background()
			org := viper.GetString("org")

			if org == "" {
				return fmt.Errorf("must specify pulumi org via flag or config file")
			}

			project := workspace.Project{
				Name:    tokens.PackageName("fh-" + config.Slug),
				Runtime: workspace.NewProjectRuntimeInfo("go", nil),
			}

			nilProgram := auto.Program(func(pCtx *pulumi.Context) error { return nil })

			workspace, err := auto.NewLocalWorkspace(ctx, nilProgram, auto.Project(project))
			if err != nil {
				return fmt.Errorf("error creating local workspace: %w", err)
			}

			stackList, err := workspace.ListStacks(ctx)
			if err != nil {
				return fmt.Errorf("failed to list available stacks: %w", err)
			}

			if len(stackList) > 0 {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Name", "Last Update", "Deployment Info", "URL"})

				for _, values := range stackList {
					stackName := auto.FullyQualifiedStackName(org, "fh", values.Name)
					stack, err := auto.SelectStack(ctx, stackName, workspace)
					if err != nil {
						return fmt.Errorf("error selecting stack")
					}

					out, err := stack.Outputs(ctx)
					if err != nil {
						return fmt.Errorf("no stack outpus found: %w", err)
					}

					var url string
					if out["address"].Value == nil {
						url = ""
					} else {
						url = fmt.Sprintf("http://%s", out["address"].Value.(string))
					}

					table.Append([]string{values.Name, values.LastUpdate, values.URL, url})
				}

				table.Render()
			} else {
				log.Info("No fh apps currently deployed")
			}
			return nil
		},
	}
	return command
}
