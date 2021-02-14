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
	"fmt"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

		c := make(chan Message)

		go func() {
			config, err := rest.InClusterConfig()
			if err != nil {
				panic(err.Error())
			}

			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}

			serverLogger := appLogger.Named("server")
			for range time.NewTicker(5 * time.Second).C {
				serverLogger.Debug("Starting List")
				pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					panic(err.Error())
				}
				if serverLogger.IsDebug() {
					for _, i := range pods.Items {
						serverLogger.Debug("Found POD", "name", i.GetName())
					}
				}
				c <- WarningMessage(fmt.Sprintf("Number of Pods: %d", len(pods.Items)))
				serverLogger.Debug("Ending List")
			}
		}()

		notifyLogger := appLogger.Named("notify")

		n := LogNotifier(notifyLogger)
		for m := range c {
			notifyLogger.Debug("received message", "content", m)
			switch m.(type) {
			case WarningMessage:
				notifyLogger.Debug("sent message downstream", "type", "warning", "content", m)
				n.Warn(m)
			case ErrorMessage:
				notifyLogger.Debug("sent message downstream", "type", "error", "content", m)
				n.Error(m)
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
