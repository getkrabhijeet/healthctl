package testsuite

import (
	"healthctl/pkg/models"

	"k8s.io/client-go/kubernetes"
)

func CheckUPF(clientset *kubernetes.Clientset) []models.ResourceCheck {
	checks := []models.ResourceCheck{}
	// checks = append(checks, CheckPods(clientset)...)
	// checks = append(checks, CheckSMFMonitor(clientset)...)
	return checks
}
