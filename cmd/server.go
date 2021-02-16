/*
Copyright © 2021 Markus Mahlberg <markus@mahlberg.io>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/velero-sentinel/sentinel/notification"

	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"

	"github.com/vmware-tanzu/velero/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		appLogger.Info("Starting server")

		appLogger.Info("Connecting to kubernetes master")

		appLogger.Info("Loading config")
		cfg, err := client.LoadConfig()
		if err != nil {
			appLogger.Error("Loading config", "error", err)
		}

		appLogger.Info("Setting up client factory")
		factory := client.NewFactory("sentinel", cfg)

		appLogger.Info("Setting up client")
		myclient, err := factory.Client()

		if err != nil {
			appLogger.Error("Setting up client", "error", err)
		}

		appLogger.Info("Setting up watcher")
		watcher, err := myclient.VeleroV1().Backups("velero").Watch(context.Background(), metav1.ListOptions{})

		if err != nil {
			appLogger.Error("Setting up watcher", "error", err)
			panic(err)
		}

		c, err := notification.Setup(viper.Sub("notification"), appLogger.Named("notifications"))

		if err != nil {
			appLogger.Error("Setting up notifications", "error", err)
			panic(err)
		}

		for {
			select {
			case evt := <-watcher.ResultChan():

				backup, ok := evt.Object.(*v1.Backup)

				if !ok {
					appLogger.Error("Non-Backup event registered", "evt", backup)
					continue
				}

				switch evt.Type {
				case watch.Added:
					appLogger.Info("Backup added", "name", backup.Name)
					continue
				case watch.Deleted:
					appLogger.Info("Backup deleted", "name", backup.Name)
					continue
				}

				switch backup.Status.Phase {
				case v1.BackupPhaseNew:
					// There IS a state "new"... However, I do not think this actually will come
					// up a lot, unless you have many potentially concurrent backup
					appLogger.Info("New backup detected", "name", backup.Name, "state", evt.Type)
				case v1.BackupPhaseCompleted:
					appLogger.Info("Backup completed", "name", backup.Name, "state", evt.Type)
				case v1.BackupPhaseDeleting:
					appLogger.Info("Backup deletion", "name", backup.Name, "state", evt.Type)
				case v1.BackupPhaseInProgress:
					appLogger.Info("Backup in progress", "name", backup.Name, "state", evt.Type)
				case v1.BackupPhasePartiallyFailed:
					c <- notification.WarningMessage{Backup: backup}
				case v1.BackupPhaseFailed:
					c <- notification.ErrorMessage{Backup: backup}
				}
			}
		}

	},
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
