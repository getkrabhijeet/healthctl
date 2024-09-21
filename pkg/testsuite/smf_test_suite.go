package testsuite

import (
	"context"
	"fmt"
	"healthctl/pkg/models"
	"os/exec"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CheckSMF(clientset *kubernetes.Clientset) []models.ResourceCheck {
	checks := []models.ResourceCheck{}
	checks = append(checks, CheckPods(clientset)...)
	checks = append(checks, CheckSMFMonitor(clientset)...)
	return checks
}

// Check functions
func CheckPods(clientset *kubernetes.Clientset) []models.ResourceCheck {
	ctx := context.TODO()
	pods, err := clientset.CoreV1().Pods("fed-smf").List(ctx, metav1.ListOptions{})
	if err != nil {
		return append([]models.ResourceCheck{}, models.ResourceCheck{
			Label:   "Pods",
			Details: "Failed to list pods in SMF namespace",
			Status:  false,
		})
	}

	deploymentStatus := make(map[string]bool)
	deploymentDetails := make(map[string][]string)

	for _, pod := range pods.Items {
		deploymentName := pod.Labels["app"]
		allContainersReady := true

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				allContainersReady = false
				break
			}
		}

		if _, exists := deploymentStatus[deploymentName]; !exists {
			deploymentStatus[deploymentName] = true
		}

		if !allContainersReady {
			deploymentStatus[deploymentName] = false
		}

		deploymentDetails[deploymentName] = append(deploymentDetails[deploymentName], pod.Name)
	}

	var checks []models.ResourceCheck
	for deployment, status := range deploymentStatus {
		checks = append(checks, models.ResourceCheck{
			Label:   deployment,
			Details: fmt.Sprintf("Deployment: %s", strings.Join(deploymentDetails[deployment], ", ")),
			Status:  status,
		})
	}

	return checks
}

func CheckSMFMonitor(clientset *kubernetes.Clientset) []models.ResourceCheck {
	ctx := context.TODO()
	pods, err := clientset.CoreV1().Pods("fed-smf").List(ctx, metav1.ListOptions{
		LabelSelector: "app=smfmonitor-app",
	})
	if err != nil || len(pods.Items) == 0 {
		return append([]models.ResourceCheck{}, models.ResourceCheck{
			Label:   "SMF Monitor",
			Details: "Failed to find smf-monitor pod",
			Status:  false,
		})
	}
	podName := pods.Items[0].Name
	cmd := []string{
		"kubectl",
		"exec",
		"-n", "fed-smf",
		podName,
		"-c", "smfmonitor",
		"--",
		"sh",
		"-c",
		"curl -s http://127.0.0.1:9090/tenv/SmfMonitorCliIf/smfmonitor-info | grep -A26 'Critical Ready Services Monitoring Status' ",
	}

	output, _ := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	var checks []models.ResourceCheck
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "ServiceName:") {
			parts := strings.Split(line, "|")
			serviceName := strings.TrimSpace(strings.Split(parts[0], ":")[1])
			currentInstances := ""
			minInstances := ""
			status := ""
			for _, part := range parts {
				if strings.Contains(part, "Current No Of Instances") || strings.Contains(part, "Current Available Servers Count") {
					currentInstances = strings.TrimSpace(strings.Split(part, ":")[2])
				}
				if strings.Contains(part, "Minimum No Of Instances Required") || strings.Contains(part, "Minimum No Of Clusters Required") || strings.Contains(part, "Minimum No Of Servers Required") {
					minInstances = strings.TrimSpace(strings.Split(part, ":")[1])
					minInstances = strings.Split(minInstances, ",")[0]
				}
				if strings.Contains(part, "Status") {
					status = strings.TrimSpace(strings.Split(part, ":")[1])
				}
			}
			checks = append(checks, models.ResourceCheck{
				Label:   serviceName,
				Details: fmt.Sprintf("Service %s: Avl: %s, Min Req: %s ", serviceName, currentInstances, minInstances),
				Status:  status == "UP",
			})
		}
	}
	return checks
}
