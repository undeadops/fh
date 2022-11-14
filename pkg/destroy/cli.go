package destroy

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/prometheus/common/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/undeadops/fh/pkg/config"
)

var (
	dryrun    bool
	directory string
	verbose   bool
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "destroy",
		Short: "Remove your application",
		Long:  "Remove your applications from Kubernetes and AWS",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			//ctx := context.Background()
			config, err := config.NewConfig(viper.GetString("values"))

			label := fmt.Sprintf("This will delete the application %s. Are you sure you wish to continue?", config.Name)

			prompt := promptui.Prompt{
				Label:     label,
				IsConfirm: true,
			}

			result, err := prompt.Run()

			if err != nil {
				fmt.Printf("User cancelled, not deleting %v\n", err)
				os.Exit(0)
			}

			log.Debug("Region: %s", config.Aws.Region)
			log.Debug("User confirmed, continuing: %s", result)
			log.Infof("Deleting application: %s", config.Name)

			return nil
		},
	}
	f := command.Flags()
	f.BoolVarP(&dryrun, "preview", "p", false, "Preview changes, dry-run mode")
	f.BoolVarP(&verbose, "verbose", "v", false, "Show output of Pulumi operations")
	f.StringVarP(&directory, "dir", "d", ".", "Path to docker context to use")

	viper.BindPFlag("stack", command.Flags().Lookup("stack"))

	cobra.MarkFlagRequired(f, "name")
	return command
}
