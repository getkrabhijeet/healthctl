package testsuite

import (
	"context"

	"healthctl/pkg/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CheckINFRA(clientset *kubernetes.Clientset) []models.ResourceCheck {
	checks := []models.ResourceCheck{
		CheckOPA(clientset),
		CheckMetallb(clientset),
		CheckKubeAddons(clientset),
		CheckFedRbac(clientset),
		CheckFedCRD(clientset),
	}
	return checks
}

// Check functions
func CheckOPA(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if OPA pod is running in fed-opa namespace
	pods, err := clientset.CoreV1().Pods("fed-opa").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"OPA", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"OPA", "No OPA pods found", false}
	}

	// Check if OPA service is up
	services, err := clientset.CoreV1().Services("fed-opa").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"OPA", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"OPA", "No OPA services found", false}
	}

	return models.ResourceCheck{"OPA", "OPA is Up", true}
}

func CheckMetallb(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if MetalLB pod is running in fed-metallb-system namespace
	pods, err := clientset.CoreV1().Pods("fed-metallb-system").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"MetalLB", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"MetalLB", "No MetalLB pods found", false}
	}

	// Check if MetalLB service is up
	services, err := clientset.CoreV1().Services("fed-metallb").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"MetalLB", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"MetalLB", "No MetalLB services found", false}
	}

	return models.ResourceCheck{"MetalLB", "MetalLB is Up", true}
}

func CheckKubeAddons(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if kube-addons pod is running in fed-kube-addons namespace
	pods, err := clientset.CoreV1().Pods("fed-kube-addons").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"KubeAddons", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"KubeAddons", "No KubeAddons pods found", false}
	}

	// Check if kube-addons service is up
	services, err := clientset.CoreV1().Services("fed-kube-addons").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"KubeAddons", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"KubeAddons", "No KubeAddons services found", false}
	}

	return models.ResourceCheck{"KubeAddons", "KubeAddons is Up", true}
}

func CheckFedRbac(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if fed-rbac pod is running in fed-rbac namespace
	pods, err := clientset.CoreV1().Pods("fed-rbac").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"FedRbac", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"FedRbac", "No Rbac pods found", false}
	}

	return models.ResourceCheck{"FedRbac", "FedRbac is Up", true}
}

func CheckFedCRD(clientset *kubernetes.Clientset) models.ResourceCheck {

	return models.ResourceCheck{"FedCRD", "FedCRD is Up", true}
}
