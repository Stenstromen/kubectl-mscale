package scale

import (
	"context"
	"fmt"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestScaleResource(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a test deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
		},
	}

	// Add the deployment to the fake clientset
	_, err := clientset.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test deployment: %v", err)
	}

	// Test scaling the deployment
	err = ScaleResource(clientset, "deployment", "test-deployment", "default", 5, 3)
	if err != nil {
		t.Fatalf("Failed to scale deployment: %v", err)
	}

	// Verify the deployment was scaled
	updatedDeployment, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), "test-deployment", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get updated deployment: %v", err)
	}

	if *updatedDeployment.Spec.Replicas != 5 {
		t.Errorf("Expected 5 replicas, got %d", *updatedDeployment.Spec.Replicas)
	}
}

func TestScaleResourceWithInvalidCurrentReplicas(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a test deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
		},
	}

	// Add the deployment to the fake clientset
	_, err := clientset.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test deployment: %v", err)
	}

	// Test scaling with invalid current replicas
	err = ScaleResource(clientset, "deployment", "test-deployment", "default", 5, 2)
	if err == nil {
		t.Error("Expected error when current replicas don't match, got nil")
	}
}

func TestScaleResourceWithNonExistentResource(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Test scaling a non-existent deployment
	err := ScaleResource(clientset, "deployment", "non-existent", "default", 5, -1)
	if err == nil {
		t.Error("Expected error when scaling non-existent deployment, got nil")
	}
}

func TestScaleResourceAcrossNamespaces(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Define test namespaces
	namespaces := []string{"default", "staging", "production"}

	// Create each namespace first
	for _, ns := range namespaces {
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}
		_, err := clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create namespace %s: %v", ns, err)
		}
	}

	// Create a deployment in each namespace
	for _, ns := range namespaces {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: ns,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(2),
			},
		}

		// Add the deployment to the fake clientset
		_, err := clientset.AppsV1().Deployments(ns).Create(context.TODO(), deployment, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create test deployment in namespace %s: %v", ns, err)
		}
	}

	// Print all deployments to verify creation
	fmt.Println("Verifying deployments were created:")
	for _, ns := range namespaces {
		deployments, err := clientset.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list deployments in namespace %s: %v", ns, err)
		}
		fmt.Printf("Namespace %s has %d deployments\n", ns, len(deployments.Items))
		for _, d := range deployments.Items {
			fmt.Printf("  - %s: %d replicas\n", d.Name, *d.Spec.Replicas)
		}
	}

	// Scale all deployments across namespaces using the new function with the fake clientset
	namespacesStr := strings.Join(namespaces, ",")
	fmt.Println("Scaling deployments across namespaces:", namespacesStr)
	err := ScaleAllResourcesWithClientset(clientset, "deployment", namespacesStr, 4, 2)
	if err != nil {
		t.Fatalf("Failed to scale deployments across namespaces: %v", err)
	}

	// Verify that deployments in each namespace were scaled correctly
	for _, ns := range namespaces {
		deployment, err := clientset.AppsV1().Deployments(ns).Get(context.TODO(), "test-deployment", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get deployment in namespace %s: %v", ns, err)
		}

		if *deployment.Spec.Replicas != 4 {
			t.Errorf("Expected 4 replicas in namespace %s, got %d", ns, *deployment.Spec.Replicas)
		}

		if deployment.Namespace != ns {
			t.Errorf("Expected deployment in namespace %s, got %s", ns, deployment.Namespace)
		}
	}
}
