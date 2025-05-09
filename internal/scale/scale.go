package scale

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeConfigPath() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	return kubeconfig
}

// ScaleFromFile scales resources defined in a YAML file
func ScaleFromFile(filename string, replicas, currentReplicas int) error {
	// Get kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", getKubeConfigPath())
	if err != nil {
		return fmt.Errorf("error building kubeconfig: %v", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	// Read the file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewYAMLOrJSONDecoder(file, 4096)
	for {
		obj := &unstructured.Unstructured{}
		if err := decoder.Decode(obj); err != nil {
			break
		}

		if obj.GetKind() == "" {
			continue
		}

		// Get the resource type and name
		resourceType := strings.ToLower(obj.GetKind())
		resourceName := obj.GetName()
		namespace := obj.GetNamespace()
		if namespace == "" {
			namespace = "default"
		}

		// Scale the resource
		if err := ScaleResource(clientset, resourceType, resourceName, namespace, replicas, currentReplicas); err != nil {
			fmt.Printf("Error scaling %s %s: %v\n", resourceType, resourceName, err)
		}
	}

	return nil
}

// ScaleFromArgs scales resources specified by command line arguments
func ScaleFromArgs(args []string, resourceType string, namespaces string, replicas, currentReplicas int) error {
	// Get kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", getKubeConfigPath())
	if err != nil {
		return fmt.Errorf("error building kubeconfig: %v", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	// Parse resource names (now supports both formats for backward compatibility)
	resourceNames := make([]string, 0)
	for _, arg := range args {
		// If it contains a slash, extract just the name part (for backward compatibility)
		if strings.Contains(arg, "/") {
			parts := strings.Split(arg, "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid resource format: %s", arg)
			}
			resourceNames = append(resourceNames, parts[1])
		} else {
			// Use the name directly
			resourceNames = append(resourceNames, arg)
		}
	}

	// Parse namespaces
	namespaceList := []string{"default"}
	if namespaces != "" {
		namespaceList = strings.Split(namespaces, ",")
	}

	// Scale each resource in each namespace
	for _, ns := range namespaceList {
		for _, name := range resourceNames {
			if err := ScaleResource(clientset, resourceType, name, ns, replicas, currentReplicas); err != nil {
				fmt.Printf("Error scaling %s %s in namespace %s: %v\n", resourceType, name, ns, err)
			}
		}
	}

	return nil
}

// ScaleResource scales a specific resource
func ScaleResource(clientset kubernetes.Interface, resourceType, name, namespace string, replicas, currentReplicas int) error {
	fmt.Printf("Scaling %s %s in namespace %s to %d replicas...\n", resourceType, name, namespace, replicas)

	switch strings.ToLower(resourceType) {
	case "deployment", "deploy", "deployments":
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting deployment: %v", err)
		}

		if currentReplicas != -1 && int(*deployment.Spec.Replicas) != currentReplicas {
			return fmt.Errorf("current replicas %d doesn't match expected %d", *deployment.Spec.Replicas, currentReplicas)
		}

		deployment.Spec.Replicas = int32Ptr(int32(replicas))
		_, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error scaling: %v", err)
		}

	case "statefulset", "sts", "statefulsets":
		statefulset, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting statefulset: %v", err)
		}

		if currentReplicas != -1 && int(*statefulset.Spec.Replicas) != currentReplicas {
			return fmt.Errorf("current replicas %d doesn't match expected %d", *statefulset.Spec.Replicas, currentReplicas)
		}

		statefulset.Spec.Replicas = int32Ptr(int32(replicas))
		_, err = clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulset, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error scaling: %v", err)
		}

	case "replicaset", "rs", "replicasets":
		replicaset, err := clientset.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting replicaset: %v", err)
		}

		if currentReplicas != -1 && int(*replicaset.Spec.Replicas) != currentReplicas {
			return fmt.Errorf("current replicas %d doesn't match expected %d", *replicaset.Spec.Replicas, currentReplicas)
		}

		replicaset.Spec.Replicas = int32Ptr(int32(replicas))
		_, err = clientset.AppsV1().ReplicaSets(namespace).Update(context.TODO(), replicaset, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error scaling: %v", err)
		}

	case "replicationcontroller", "rc", "replicationcontrollers":
		rc, err := clientset.CoreV1().ReplicationControllers(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting replicationcontroller: %v", err)
		}

		if currentReplicas != -1 && int(*rc.Spec.Replicas) != currentReplicas {
			return fmt.Errorf("current replicas %d doesn't match expected %d", *rc.Spec.Replicas, currentReplicas)
		}

		rc.Spec.Replicas = int32Ptr(int32(replicas))
		_, err = clientset.CoreV1().ReplicationControllers(namespace).Update(context.TODO(), rc, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error scaling: %v", err)
		}

	case "job", "jobs":
		job, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting job: %v", err)
		}

		if currentReplicas != -1 && int(*job.Spec.Parallelism) != currentReplicas {
			return fmt.Errorf("current parallelism %d doesn't match expected %d", *job.Spec.Parallelism, currentReplicas)
		}

		job.Spec.Parallelism = int32Ptr(int32(replicas))
		_, err = clientset.BatchV1().Jobs(namespace).Update(context.TODO(), job, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error scaling: %v", err)
		}

	case "cronjob", "cj", "cronjobs":
		cronjob, err := clientset.BatchV1().CronJobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting cronjob: %v", err)
		}

		if currentReplicas != -1 && int(*cronjob.Spec.JobTemplate.Spec.Parallelism) != currentReplicas {
			return fmt.Errorf("current parallelism %d doesn't match expected %d", *cronjob.Spec.JobTemplate.Spec.Parallelism, currentReplicas)
		}

		cronjob.Spec.JobTemplate.Spec.Parallelism = int32Ptr(int32(replicas))
		_, err = clientset.BatchV1().CronJobs(namespace).Update(context.TODO(), cronjob, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error scaling: %v", err)
		}

	case "horizontalpodautoscaler", "hpa", "horizontalpodautoscalers":
		hpa, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting horizontalpodautoscaler: %v", err)
		}

		if currentReplicas != -1 && int(*hpa.Spec.MinReplicas) != currentReplicas {
			return fmt.Errorf("current min replicas %d doesn't match expected %d", *hpa.Spec.MinReplicas, currentReplicas)
		}

		hpa.Spec.MinReplicas = int32Ptr(int32(replicas))
		hpa.Spec.MaxReplicas = int32(replicas)
		_, err = clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).Update(context.TODO(), hpa, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error scaling: %v", err)
		}

	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	fmt.Printf("Successfully scaled %s %s to %d replicas\n", resourceType, name, replicas)
	return nil
}

// Helper function to create int32 pointer
func int32Ptr(i int32) *int32 {
	return &i
}

// ScaleAllResources scales all resources of the specified type in the given namespaces
func ScaleAllResources(resourceType, namespaces string, replicas, currentReplicas int) error {
	// Get kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", getKubeConfigPath())
	if err != nil {
		return fmt.Errorf("error building kubeconfig: %v", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	return ScaleAllResourcesWithClientset(clientset, resourceType, namespaces, replicas, currentReplicas)
}

// ScaleAllResourcesWithClientset scales all resources of the specified type in the given namespaces using the provided clientset
func ScaleAllResourcesWithClientset(clientset kubernetes.Interface, resourceType, namespaces string, replicas, currentReplicas int) error {
	// Parse namespaces
	namespaceList := []string{"default"}
	if namespaces != "" {
		namespaceList = strings.Split(namespaces, ",")
	}

	// Scale all resources of the specified type in each namespace
	for _, ns := range namespaceList {
		switch strings.ToLower(resourceType) {
		case "deployment", "deploy", "deployments":
			deployments, err := clientset.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Error listing deployments in namespace %s: %v\n", ns, err)
				continue
			}

			if len(deployments.Items) == 0 {
				fmt.Printf("No deployments found in namespace %s\n", ns)
				continue
			}

			for _, deployment := range deployments.Items {
				if err := ScaleResource(clientset, "deployment", deployment.Name, ns, replicas, currentReplicas); err != nil {
					fmt.Printf("Error scaling deployment %s in namespace %s: %v\n", deployment.Name, ns, err)
				}
			}

		case "statefulset", "sts", "statefulsets":
			statefulsets, err := clientset.AppsV1().StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Error listing statefulsets in namespace %s: %v\n", ns, err)
				continue
			}

			if len(statefulsets.Items) == 0 {
				fmt.Printf("No statefulsets found in namespace %s\n", ns)
				continue
			}

			for _, statefulset := range statefulsets.Items {
				if err := ScaleResource(clientset, "statefulset", statefulset.Name, ns, replicas, currentReplicas); err != nil {
					fmt.Printf("Error scaling statefulset %s in namespace %s: %v\n", statefulset.Name, ns, err)
				}
			}

		case "replicaset", "rs", "replicasets":
			replicasets, err := clientset.AppsV1().ReplicaSets(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Error listing replicasets in namespace %s: %v\n", ns, err)
				continue
			}

			if len(replicasets.Items) == 0 {
				fmt.Printf("No replicasets found in namespace %s\n", ns)
				continue
			}

			for _, replicaset := range replicasets.Items {
				if err := ScaleResource(clientset, "replicaset", replicaset.Name, ns, replicas, currentReplicas); err != nil {
					fmt.Printf("Error scaling replicaset %s in namespace %s: %v\n", replicaset.Name, ns, err)
				}
			}

		case "replicationcontroller", "rc", "replicationcontrollers":
			rcs, err := clientset.CoreV1().ReplicationControllers(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Error listing replicationcontrollers in namespace %s: %v\n", ns, err)
				continue
			}

			if len(rcs.Items) == 0 {
				fmt.Printf("No replicationcontrollers found in namespace %s\n", ns)
				continue
			}

			for _, rc := range rcs.Items {
				if err := ScaleResource(clientset, "replicationcontroller", rc.Name, ns, replicas, currentReplicas); err != nil {
					fmt.Printf("Error scaling replicationcontroller %s in namespace %s: %v\n", rc.Name, ns, err)
				}
			}

		case "job", "jobs":
			jobs, err := clientset.BatchV1().Jobs(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Error listing jobs in namespace %s: %v\n", ns, err)
				continue
			}

			if len(jobs.Items) == 0 {
				fmt.Printf("No jobs found in namespace %s\n", ns)
				continue
			}

			for _, job := range jobs.Items {
				if err := ScaleResource(clientset, "job", job.Name, ns, replicas, currentReplicas); err != nil {
					fmt.Printf("Error scaling job %s in namespace %s: %v\n", job.Name, ns, err)
				}
			}

		case "cronjob", "cj", "cronjobs":
			cronjobs, err := clientset.BatchV1().CronJobs(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Error listing cronjobs in namespace %s: %v\n", ns, err)
				continue
			}

			if len(cronjobs.Items) == 0 {
				fmt.Printf("No cronjobs found in namespace %s\n", ns)
				continue
			}

			for _, cronjob := range cronjobs.Items {
				if err := ScaleResource(clientset, "cronjob", cronjob.Name, ns, replicas, currentReplicas); err != nil {
					fmt.Printf("Error scaling cronjob %s in namespace %s: %v\n", cronjob.Name, ns, err)
				}
			}

		case "horizontalpodautoscaler", "hpa", "horizontalpodautoscalers":
			hpas, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Error listing horizontalpodautoscalers in namespace %s: %v\n", ns, err)
				continue
			}

			if len(hpas.Items) == 0 {
				fmt.Printf("No horizontalpodautoscalers found in namespace %s\n", ns)
				continue
			}

			for _, hpa := range hpas.Items {
				if err := ScaleResource(clientset, "horizontalpodautoscaler", hpa.Name, ns, replicas, currentReplicas); err != nil {
					fmt.Printf("Error scaling horizontalpodautoscaler %s in namespace %s: %v\n", hpa.Name, ns, err)
				}
			}

		default:
			return fmt.Errorf("unsupported resource type: %s", resourceType)
		}
	}

	return nil
}
