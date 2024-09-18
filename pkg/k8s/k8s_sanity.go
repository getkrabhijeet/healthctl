package k8s

import (
    "context"
    "fmt"
    "path/filepath"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceCheck struct {
    Label   string
    Details string
    Status  bool
}

func main() {
    // Load kubeconfig
    kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        fmt.Println("Error loading kubeconfig:", err)
        return
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        fmt.Println("Error creating Kubernetes client:", err)
        return
    }

    resourceChecks := []ResourceCheck{
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

    // Display the structured results
    printResults(resourceChecks)
}

func printResults(resourceChecks []ResourceCheck) {
    fmt.Println("Health Check Summary:")
    fmt.Println("---------------------")
    for _, check := range resourceChecks {
        fmt.Printf("%s:\n", check.Label)
        fmt.Printf("  Details: %s\n", check.Details)
        fmt.Printf("  Status: %t\n\n", check.Status)
    }
}

// Check functions
func checkNodes(clientset *kubernetes.Clientset) ResourceCheck {
    nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Nodes", "Error fetching nodes", false}
    }

    nodeNames := make([]string, len(nodes.Items))
    for i, node := range nodes.Items {
        nodeNames[i] = node.Name
    }

    return ResourceCheck{"Nodes", fmt.Sprintf("%d nodes found: %s", len(nodes.Items), nodeNames), true}
}

func checkPods(clientset *kubernetes.Clientset) ResourceCheck {
    pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Pods", "Error fetching pods", false}
    }

    totalPods := len(pods.Items)
    healthyPods := 0

    for _, pod := range pods.Items {
        if pod.Status.Phase == "Running" || pod.Status.Phase == "Succeeded" {
            healthyPods++
        }
    }

    allPodsHealthy := healthyPods == totalPods
    details := fmt.Sprintf("Total: %d, Healthy: %d. Status: %s", totalPods, healthyPods,
        getPodsHealthMessage(totalPods, healthyPods))
    return ResourceCheck{"Pods", details, allPodsHealthy}
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

func checkPVs(clientset *kubernetes.Clientset) ResourceCheck {
    pvs, err := clientset.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Persistent Volumes", "Error fetching persistent volumes", false}
    }

    count := len(pvs.Items)
    details := fmt.Sprintf("Total: %d", count)
    if count == 0 {
        details = "No persistent volumes are available."
        return ResourceCheck{"Persistent Volumes", details, false}
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

    return ResourceCheck{"Persistent Volumes", details, allBound}
}

func checkPVCs(clientset *kubernetes.Clientset) ResourceCheck {
    pvcs, err := clientset.CoreV1().PersistentVolumeClaims("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Persistent Volume Claims", "Error fetching persistent volume claims", false}
    }

    count := len(pvcs.Items)
    details := fmt.Sprintf("Total: %d", count)
    if count == 0 {
        details = "No persistent volume claims are available."
        return ResourceCheck{"Persistent Volume Claims", details, false}
    }

    return ResourceCheck{"Persistent Volume Claims", details, true}
}

func checkServices(clientset *kubernetes.Clientset) ResourceCheck {
    services, err := clientset.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Services", "Error fetching services", false}
    }
    count := len(services.Items)
    details := fmt.Sprintf("Total: %d", count)
    if count == 0 {
        details = "No services are available."
    }
    return ResourceCheck{"Services", details, count > 0}
}

func checkDeployments(clientset *kubernetes.Clientset) ResourceCheck {
    deployments, err := clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Deployments", "Error fetching deployments", false}
    }

    count := len(deployments.Items)
    details := fmt.Sprintf("Total: %d", count)
    if count == 0 {
        details = "No deployments are available."
        return ResourceCheck{"Deployments", details, false}
    }

    allHealthy := true
    for _, deploy := range deployments.Items {
        if *deploy.Spec.Replicas != deploy.Status.ReadyReplicas {
            allHealthy = false
            break
        }
    }

    if allHealthy {
        details = "All deployments are healthy."
    } else {
        details = "Some deployments are not healthy."
    }

    return ResourceCheck{"Deployments", details, allHealthy}
}

func checkReplicaSets(clientset *kubernetes.Clientset) ResourceCheck {
    replicasets, err := clientset.AppsV1().ReplicaSets("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Replica Sets", "Error fetching replica sets", false}
    }

    count := len(replicasets.Items)
    details := fmt.Sprintf("Total: %d", count)
    if count == 0 {
        details = "No replica sets are available."
        return ResourceCheck{"Replica Sets", details, false}
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

    return ResourceCheck{"Replica Sets", details, allHealthy}
}

func checkEvents(clientset *kubernetes.Clientset) ResourceCheck {
    events, err := clientset.CoreV1().Events("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Events", "Error fetching events", false}
    }

    count := len(events.Items)
    details := fmt.Sprintf("Total: %d", count)
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
            details = fmt.Sprintf("Warning events found: %s", errorEvents)
        } else {
            details = "No critical issues found in events."
        }
    }
    return ResourceCheck{"Events", details, count == 0}
}

func checkIngresses(clientset *kubernetes.Clientset) ResourceCheck {
    ingresses, err := clientset.NetworkingV1().Ingresses("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Ingresses", "Error fetching ingresses", false}
    }

    count := len(ingresses.Items)
    details := fmt.Sprintf("Total: %d", count)
    if count == 0 {
        details = "No ingresses are available."
    }
    return ResourceCheck{"Ingresses", details, count > 0}
}

func checkDaemonSets(clientset *kubernetes.Clientset) ResourceCheck {
    daemonsets, err := clientset.AppsV1().DaemonSets("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Daemon Sets", "Error fetching daemon sets", false}
    }

    count := len(daemonsets.Items)
    details := fmt.Sprintf("Total: %d", count)
    if count == 0 {
        details = "No daemon sets are available."
        return ResourceCheck{"Daemon Sets", details, false}
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

    return ResourceCheck{"Daemon Sets", details, allHealthy}
}

func checkStatefulSets(clientset *kubernetes.Clientset) ResourceCheck {
 statefulsets, err := clientset.AppsV1().StatefulSets("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        return ResourceCheck{"Stateful Sets", "Error fetching stateful sets", false}
    }

    count := len(statefulsets.Items)
    details := fmt.Sprintf("Total: %d", count)
    if count == 0 {
        details = "No stateful sets are available."
        return ResourceCheck{"Stateful Sets", details, false}
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

    return ResourceCheck{"Stateful Sets", details, allHealthy}
}
