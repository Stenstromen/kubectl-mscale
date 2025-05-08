package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stenstromen/kubectl-mscale/internal/scale"
)

var (
	replicas        int
	namespaces      string
	filename        string
	currentReplicas int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubectl-mscale",
	Short: "Scale resources across multiple namespaces",
	Long:  `A kubectl plugin for scaling resources across multiple namespaces.`,
	Example: `  # Scale a deployment named 'nginx' to 3 replicas across multiple namespaces
  kubectl-mscale deployment/nginx --replicas=3 -n default,staging,production

  # Scale multiple deployments to 0 replicas across multiple namespaces
  kubectl-mscale deployment/nginx deployment/redis --replicas=0 -n default,staging,production

  # Scale a statefulset named 'mysql' to 2 replicas across multiple namespaces
  kubectl-mscale statefulset/mysql --replicas=2 -n default,staging,production

  # Scale a replicaset named 'web' to 5 replicas across multiple namespaces
  kubectl-mscale replicaset/web --replicas=5 -n default,staging,production`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	scaleCmd := &cobra.Command{
		Use:   "deployment|sts|rs|rc|job|cj|hpa",
		Short: "Scale a specific resource type",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if filename != "" {
				return scale.ScaleFromFile(filename, replicas, currentReplicas)
			}
			return scale.ScaleFromArgs(args, namespaces, replicas, currentReplicas)
		},
	}

	scaleCmd.Flags().IntVar(&replicas, "replicas", 0, "Number of replicas")
	scaleCmd.Flags().StringVarP(&namespaces, "namespace", "n", "", "Comma-separated list of namespaces")
	scaleCmd.Flags().StringVarP(&filename, "filename", "f", "", "Filename, directory, or URL to files to use to scale the resource")
	scaleCmd.Flags().IntVar(&currentReplicas, "current-replicas", -1, "Precondition for current size. Requires that the current size of the resource match this value in order to scale")
	scaleCmd.MarkFlagRequired("replicas")

	rootCmd.AddCommand(scaleCmd)
}
