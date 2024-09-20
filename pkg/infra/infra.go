package infra

import (
	"context"

	"healthctl/pkg/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CheckINFRA(clientset *kubernetes.Clientset) []k8s.ResourceCheck {
	checks := []k8s.ResourceCheck{
		CheckOPA(clientset),
		CheckMetallb(clientset),
		CheckKubeAddons(clientset),
		CheckFedRbac(clientset),
		CheckFedCRD(clientset),
	}
	return checks
}

// Check functions
func CheckOPA(clientset *kubernetes.Clientset) k8s.ResourceCheck {

	// Check if OPA pod is running in fed-opa namespace
	pods, err := clientset.CoreV1().Pods("fed-opa").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return k8s.ResourceCheck{"OPA", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return k8s.ResourceCheck{"OPA", "No OPA pods found", false}
	}

	// Check if OPA service is up
	services, err := clientset.CoreV1().Services("fed-opa").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return k8s.ResourceCheck{"OPA", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return k8s.ResourceCheck{"OPA", "No OPA services found", false}
	}

	return k8s.ResourceCheck{"OPA", "OPA is Up", true}
}

func CheckMetallb(clientset *kubernetes.Clientset) k8s.ResourceCheck {

	// Check if MetalLB pod is running in fed-metallb-system namespace
	pods, err := clientset.CoreV1().Pods("fed-metallb-system").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return k8s.ResourceCheck{"MetalLB", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return k8s.ResourceCheck{"MetalLB", "No MetalLB pods found", false}
	}

	// Check if MetalLB service is up
	services, err := clientset.CoreV1().Services("fed-metallb").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return k8s.ResourceCheck{"MetalLB", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return k8s.ResourceCheck{"MetalLB", "No MetalLB services found", false}
	}

	return k8s.ResourceCheck{"MetalLB", "MetalLB is Up", true}
}

func CheckKubeAddons(clientset *kubernetes.Clientset) k8s.ResourceCheck {

	// Check if kube-addons pod is running in fed-kube-addons namespace
	pods, err := clientset.CoreV1().Pods("fed-kube-addons").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return k8s.ResourceCheck{"KubeAddons", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return k8s.ResourceCheck{"KubeAddons", "No KubeAddons pods found", false}
	}

	// Check if kube-addons service is up
	services, err := clientset.CoreV1().Services("fed-kube-addons").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return k8s.ResourceCheck{"KubeAddons", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return k8s.ResourceCheck{"KubeAddons", "No KubeAddons services found", false}
	}

	return k8s.ResourceCheck{"KubeAddons", "KubeAddons is Up", true}
}

func CheckFedRbac(clientset *kubernetes.Clientset) k8s.ResourceCheck {

	// Check if fed-rbac pod is running in fed-rbac namespace
	pods, err := clientset.CoreV1().Pods("fed-rbac").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return k8s.ResourceCheck{"FedRbac", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return k8s.ResourceCheck{"FedRbac", "No Rbac pods found", false}
	}

	return k8s.ResourceCheck{"FedRbac", "FedRbac is Up", true}
}

func CheckFedCRD(clientset *kubernetes.Clientset) k8s.ResourceCheck {

	return k8s.ResourceCheck{"FedCRD", "FedCRD is Up", true}
}
