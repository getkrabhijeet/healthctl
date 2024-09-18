package k8s

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

type K8sClient struct {
	Client *kubernetes.Clientset
}

func GetClustersFromKubeConfig() *clientcmdapi.Config {
	// Get clusters from kubeconfig
	var kubeconfig1 *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig1 = flag.String("kubeconfig1", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig1 = flag.String("kubeconfig1", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.LoadFromFile(*kubeconfig1)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func CreateK8sClientSet() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return kubernetes.NewForConfig(config)
}

func NewK8sClient() (*K8sClient, error) {
	client, err := CreateK8sClientSet()
	if err != nil {
		return nil, err
	}
	return &K8sClient{Client: client}, nil
}

// GetClusterInfo returns the cluster version
func (kc *K8sClient) GetClusterInfo() (string, error) {
	clusterVersion, err := kc.Client.Discovery().ServerVersion()

	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Cluster version: %s\n", clusterVersion), nil
}

// Set context for the client
func (kc *K8sClient) SetContext(config *clientcmdapi.Config, contextToSwitch string) {
	for _, contexts := range config.Contexts {
		if contexts.Cluster == contextToSwitch {
			config.CurrentContext = contextToSwitch
		}
	}
}

// GetClusterNodes returns the cluster nodes
func (kc *K8sClient) GetClusterNodes() []int {
	nodes, _ := kc.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	masterNodes := 0
	workerNodes := 0
	for _, node := range nodes.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			masterNodes++
		} else {
			workerNodes++
		}
	}
	return []int{masterNodes, workerNodes}
}

// GetClusterNamespaces returns the cluster namespaces
func (kc *K8sClient) GetClusterNamespaces() ([]string, error) {
	var namespaceList []string
	namespaces, err := kc.Client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaces.Items {
		namespaceList = append(namespaceList, namespace.Name)
	}
	return namespaceList, nil
}

func GetAPIResources(client *kubernetes.Clientset) {
	// Get all resources in the cluster
	resources, err := client.Discovery().ServerPreferredResources()
	if err != nil {
		panic(err.Error())
	}
	for _, resource := range resources {
		fmt.Printf("Resource: %s\n", resource.GroupVersion)
		for _, apiResource := range resource.APIResources {
			fmt.Printf("  Name: %s\n", apiResource.Name)
			fmt.Printf("  Namespaced: %t\n", apiResource.Namespaced)
			fmt.Printf("  Kind: %s\n", apiResource.Kind)
			fmt.Printf("  Verbs: %s\n", apiResource.Verbs)
		}
	}

}

type TestStatus struct {
	Status bool
	Info   string
	Error  error
}

// Check if all nodes are ready
func (kc *K8sClient) CheckNodes() TestStatus {
	nodes, err := kc.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return TestStatus{
			Status: false,
			Info:   "",
			Error:  err}
	}
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" {
				return TestStatus{
					Status: false,
					Info:   fmt.Sprintf("Node %s is not ready\n", node.Name),
					Error:  nil}
			}
		}
	}
	return TestStatus{
		Status: true,
		Info:   "All nodes are ready",
		Error:  nil}
}

// Check if all pods are running
func (kc *K8sClient) CheckPods() TestStatus {
	pods, err := kc.Client.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return TestStatus{
			Status: false,
			Info:   "",
			Error:  err}
	}
	notRunningCount := 0
	for _, pod := range pods.Items {

		if pod.Status.Phase != "Running" {
			notRunningCount++
		}
	}
	if notRunningCount > 0 {
		return TestStatus{
			Status: false,
			Info:   fmt.Sprintf("%d pods are in not running state\n", notRunningCount),
			Error:  nil}
	}
	return TestStatus{
		Status: true,
		Info:   "All pods are running",
		Error:  nil}
}

// Check Events on the cluster
func (kc *K8sClient) CheckEvents() TestStatus {
	events, err := kc.Client.CoreV1().Events("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return TestStatus{
			Status: false,
			Info:   "",
			Error:  err}
	}
	warningEventsCount := 0
	errorEventsCount := 0

	for _, event := range events.Items {
		if event.Type == "Warning" {
			warningEventsCount++
		}
		if event.Type == "Error" {
			errorEventsCount++
		}
	}
	if warningEventsCount > 0 || errorEventsCount > 0 {
		return TestStatus{
			Status: false,
			Info:   fmt.Sprintf("%d warning events and %d error events found\n", warningEventsCount, errorEventsCount),
			Error:  nil}
	}

	return TestStatus{
		Status: true,
		Info:   "No failed events found",
		Error:  nil}
}
