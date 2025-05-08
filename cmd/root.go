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
	all             bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubectl-mscale",
	Short: "Scale resources across multiple namespaces",
	Long:  `A kubectl plugin for scaling resources across multiple namespaces.`,
	Example: `  # Scale all deployments to 3 replicas across multiple namespaces
  kubectl-mscale deployment --replicas=3 -n default,staging,production

  # Scale a deployment named 'nginx' to 0 replicas across multiple namespaces
  kubectl-mscale deployment nginx --replicas=0 -n default,staging,production
  
  # Scale ALL deployments to 0 replicas across all namespaces
  kubectl-mscale deployment --replicas=0 --all
  
  # Scale resources defined in a YAML file
  kubectl-mscale statefulset --filename=statefulset.yaml --replicas=3`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	createScaleCommand("deployment", "deployment", "deploy", "deployments")
	createScaleCommand("statefulset", "statefulset", "sts", "statefulsets")
	createScaleCommand("replicaset", "replicaset", "rs", "replicasets")
	createScaleCommand("replicationcontroller", "replicationcontroller", "rc", "replicationcontrollers")
	createScaleCommand("job", "job", "jobs")
	createScaleCommand("cronjob", "cronjob", "cj", "cronjobs")
	createScaleCommand("horizontalpodautoscaler", "horizontalpodautoscaler", "hpa", "horizontalpodautoscalers")
}

// createScaleCommand creates a new scale command with the given name and aliases
func createScaleCommand(use string, resourceType string, aliases ...string) {
	scaleCmd := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   fmt.Sprintf("Scale %s across multiple namespaces", use),
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if filename != "" {
				return scale.ScaleFromFile(filename, replicas, currentReplicas)
			}

			// If --all flag is set or no args are provided, scale all resources of this type
			if all || len(args) == 0 {
				return scale.ScaleAllResources(resourceType, namespaces, replicas, currentReplicas)
			}

			return scale.ScaleFromArgs(args, resourceType, namespaces, replicas, currentReplicas)
		},
	}

	scaleCmd.Flags().IntVar(&replicas, "replicas", 0, "Number of replicas")
	scaleCmd.Flags().StringVarP(&namespaces, "namespace", "n", "", "Comma-separated list of namespaces")
	scaleCmd.Flags().StringVarP(&filename, "filename", "f", "", "Filename, directory, or URL to files to use to scale the resource")
	scaleCmd.Flags().IntVar(&currentReplicas, "current-replicas", -1, "Precondition for current size. Requires that the current size of the resource match this value in order to scale")
	scaleCmd.Flags().BoolVar(&all, "all", false, "Scale all resources of the specified type in the given namespaces")
	scaleCmd.MarkFlagRequired("replicas")

	rootCmd.AddCommand(scaleCmd)
}
