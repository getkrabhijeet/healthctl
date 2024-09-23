package testsuite

import (
	"context"
	"fmt"

	"healthctl/pkg/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CheckK8s(clientset *kubernetes.Clientset) []models.ResourceCheck {

	checks := []models.ResourceCheck{
		checkNodes(clientset),
		checkPods(clientset),
		checkPVs(clientset),
		checkPVCs(clientset),
		checkServices(clientset),
		checkDeployments(clientset),
		checkReplicaSets(clientset),
		checkEvents(clientset),
		checkIngresses(clientset),
		checkDaemonSets(clientset),
		checkStatefulSets(clientset),
	}

	return checks
}

// Check functions
func checkNodes(clientset *kubernetes.Clientset) models.ResourceCheck {
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Nodes", Details: "Error fetching nodes", Status: false}
	}

	nodeNames := make([]string, len(nodes.Items))
	for i, node := range nodes.Items {
		nodeNames[i] = node.Name
	}

	return models.ResourceCheck{Label: "Nodes", Details: fmt.Sprint("Number of nodes : ", len(nodes.Items)), Status: true}
}

func checkPods(clientset *kubernetes.Clientset) models.ResourceCheck {
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Pods", Details: "Error fetching pods", Status: false}
	}

	totalPods := len(pods.Items)
	healthyPods := 0

	for _, pod := range pods.Items {
		if pod.Status.Phase == "Running" || pod.Status.Phase == "Succeeded" {
			healthyPods++
		}
	}
	details := fmt.Sprintf("Total: %d, Healthy: %d. Status: %s", totalPods, healthyPods,
		getPodsHealthMessage(totalPods, healthyPods))
	return models.ResourceCheck{Label: "Pods", Details: details, Status: healthyPods == totalPods}
}

func getPodsHealthMessage(total int, healthy int) string {
	if total == 0 {
		return "No pods are available."
	}
	if healthy == total {
		return "All pods are healthy."
	}
	return fmt.Sprintf("%d out of %d pods are healthy.", healthy, total)
}

func checkPVs(clientset *kubernetes.Clientset) models.ResourceCheck {
	pvs, err := clientset.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Persistent Volumes", Details: "Error fetching persistent volumes", Status: false}
	}

	count := len(pvs.Items)
	details := fmt.Sprintf("Total: %d", count)
	if count == 0 {
		details = "No persistent volumes are available."
		return models.ResourceCheck{Label: "Persistent Volumes", Details: details, Status: false}
	}

	allBound := true
	for _, pv := range pvs.Items {
		if pv.Status.Phase != "Bound" {
			allBound = false
			break
		}
	}

	if allBound {
		details = "All persistent volumes are bound."
	} else {
		details = "Some persistent volumes are not bound."
	}

	return models.ResourceCheck{Label: "Persistent Volumes", Details: details, Status: allBound}
}

func checkPVCs(clientset *kubernetes.Clientset) models.ResourceCheck {
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Persistent Volume Claims", Details: "Error fetching persistent volume claims", Status: false}
	}

	count := len(pvcs.Items)
	details := fmt.Sprintf("Count of PVC: %d", count)
	if count == 0 {
		details = "No persistent volume claims are available."
		return models.ResourceCheck{Label: "Persistent Volume Claims", Details: details, Status: false}
	}

	return models.ResourceCheck{Label: "Persistent Volume Claims", Details: details, Status: true}
}

func checkServices(clientset *kubernetes.Clientset) models.ResourceCheck {
	services, err := clientset.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Services", Details: "Error fetching services", Status: false}
	}
	count := len(services.Items)
	details := fmt.Sprintf("Count of services: %d", count)
	if count == 0 {
		details = "No services are available."
	}
	return models.ResourceCheck{Label: "Services", Details: details, Status: count > 0}
}

func checkDeployments(clientset *kubernetes.Clientset) models.ResourceCheck {
	deployments, err := clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Deployments", Details: "Error fetching deployments", Status: false}
	}

	if len(deployments.Items) == 0 {
		return models.ResourceCheck{Label: "Deployments", Details: "No deployments are available.", Status: false}
	}

	allHealthy := true
	for _, deploy := range deployments.Items {
		if *deploy.Spec.Replicas != deploy.Status.ReadyReplicas {
			allHealthy = false
			break
		}
	}
	var details string
	if allHealthy {
		details = "All deployments are healthy."
	} else {
		details = "Some deployments are not healthy."
	}

	return models.ResourceCheck{Label: "Deployments", Details: details, Status: allHealthy}
}

func checkReplicaSets(clientset *kubernetes.Clientset) models.ResourceCheck {
	replicasets, err := clientset.AppsV1().ReplicaSets("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Replica Sets", Details: "Error fetching replica sets", Status: false}
	}

	count := len(replicasets.Items)
	details := fmt.Sprintf("Total: %d", count)
	if count == 0 {
		details = "No replica sets are available."
		return models.ResourceCheck{Label: "Replica Sets", Details: details, Status: false}
	}

	allHealthy := true
	for _, rs := range replicasets.Items {
		if *rs.Spec.Replicas != rs.Status.ReadyReplicas {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		details = "All replica sets are healthy."
	} else {
		details = "Some replica sets are not healthy."
	}

	return models.ResourceCheck{Label: "Replica Sets", Details: details, Status: allHealthy}
}

func checkEvents(clientset *kubernetes.Clientset) models.ResourceCheck {
	events, err := clientset.CoreV1().Events("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Events", Details: "Error fetching events", Status: false}
	}

	count := len(events.Items)
	details := fmt.Sprintf("Count of Events: %d", count)
	if count == 0 {
		details = "No errors found in events."
	} else {
		errorEvents := []string{}
		for _, event := range events.Items {
			if event.Type == "Warning" {
				errorEvents = append(errorEvents, event.Reason)
			}
		}
		if len(errorEvents) > 0 {
			details = fmt.Sprintf("Warning events found: %d", len(errorEvents))
		} else {
			details = "No critical issues found in events."
		}
	}
	return models.ResourceCheck{Label: "Events", Details: details, Status: count == 0}
}

func checkIngresses(clientset *kubernetes.Clientset) models.ResourceCheck {
	ingresses, err := clientset.NetworkingV1().Ingresses("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Ingresses", Details: "Error fetching ingresses", Status: false}
	}

	count := len(ingresses.Items)
	details := fmt.Sprintf("Total: %d", count)
	if count == 0 {
		details = "No ingresses are available."
	}
	return models.ResourceCheck{Label: "Ingresses", Details: details, Status: count > 0}
}

func checkDaemonSets(clientset *kubernetes.Clientset) models.ResourceCheck {
	daemonsets, err := clientset.AppsV1().DaemonSets("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Daemon Sets", Details: "Error fetching daemon sets", Status: false}
	}

	count := len(daemonsets.Items)
	details := fmt.Sprintf("Count of Daemonsets: %d", count)
	if count == 0 {
		details = "No daemon sets are available."
		return models.ResourceCheck{Label: "Daemon Sets", Details: details, Status: false}
	}

	allHealthy := true
	for _, ds := range daemonsets.Items {
		if ds.Status.DesiredNumberScheduled != ds.Status.CurrentNumberScheduled {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		details = "All daemon sets are healthy."
	} else {
		details = "Some daemon sets are not healthy."
	}

	return models.ResourceCheck{Label: "Daemon Sets", Details: details, Status: allHealthy}
}

func checkStatefulSets(clientset *kubernetes.Clientset) models.ResourceCheck {
	statefulsets, err := clientset.AppsV1().StatefulSets("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{Label: "Stateful Sets", Details: "Error fetching stateful sets", Status: false}
	}

	count := len(statefulsets.Items)
	details := fmt.Sprintf("Count of StatefulSets: %d", count)
	if count == 0 {
		details = "No stateful sets are available."
		return models.ResourceCheck{Label: "Stateful Sets", Details: details, Status: false}
	}

	allHealthy := true
	for _, ss := range statefulsets.Items {
		if *ss.Spec.Replicas != ss.Status.ReadyReplicas {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		details = "All stateful sets are healthy."
	} else {
		details = "Some stateful sets are not healthy."
	}

	return models.ResourceCheck{Label: "Stateful Sets", Details: details, Status: allHealthy}
}
