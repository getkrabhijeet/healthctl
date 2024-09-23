package k8s

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"bytes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"

	v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/client-go/tools/remotecommand"
)

var kubeconfig *string
var contextFlag *string

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		contextFlag = flag.String("context", "", "")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		contextFlag = flag.String("context", "", "")
	}
}

type K8sClient struct {
	Client        *kubernetes.Clientset
	DynamicClient dynamic.Interface
}

func GetClustersFromKubeConfig() *clientcmdapi.Config {
	flag.Parse()
	config, err := clientcmd.LoadFromFile(*kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func CreateK8sClientSet() (*kubernetes.Clientset, error) {

	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return kubernetes.NewForConfig(config)
}

func CreateDynamicClientSet() (dynamic.Interface, error) {
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return dynamic.NewForConfig(config)
}

func NewK8sClient() (*K8sClient, error) {
	client, err := CreateK8sClientSet()
	if err != nil {
		return nil, err
	}

	dynamicClient, err := CreateDynamicClientSet()
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		Client:        client,
		DynamicClient: dynamicClient,
	}, nil
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

	for key, contexts := range config.Contexts {
		if contexts.Cluster == contextToSwitch {
			config.CurrentContext = key
		}
	}
	// Write the config back to the kubeconfig file
	clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), *config, false)
	//load the client again with config
	client, err := CreateK8sClientSet()
	if err != nil {
		panic(err.Error())
	}
	kc.Client = client
	dclient, err := CreateDynamicClientSet()
	if err != nil {
		panic(err.Error())
	}
	kc.DynamicClient = dclient

}

func (kc *K8sClient) GetCurrentContext() string {
	//get context from kubeconfig
	config := GetClustersFromKubeConfig()
	return config.CurrentContext
}

func (kc *K8sClient) GetCurrentCluster() string {
	cluster := ""
	config := GetClustersFromKubeConfig()
	for key, contexts := range config.Contexts {
		if key == config.CurrentContext {
			cluster = contexts.Cluster
		}
	}
	return cluster
}

// Get Cluster Nodes names returns the cluster node name as a list
func (kc *K8sClient) GetClusterNodesName() []string {
	nodes, _ := kc.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	nodeNames := []string{}
	for _, node := range nodes.Items {
		nodeNames = append(nodeNames, node.Name)
	}
	return nodeNames
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
func (kc *K8sClient) GetClusterNamespaces() []string {
	var namespaceList []string
	namespaces, err := kc.Client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	for _, namespace := range namespaces.Items {
		namespaceList = append(namespaceList, namespace.Name)
	}
	return namespaceList
}

func (kc *K8sClient) GetPods(namespace string) []string {
	var podList []string
	pods, err := kc.Client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	for _, pod := range pods.Items {
		podList = append(podList, pod.Name)
	}
	return podList
}

func (kc *K8sClient) GetContainers(pod string) []string {
	var containerList []string
	pods, err := kc.Client.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", pod),
	})
	if err != nil {
		return nil
	}
	for _, container := range pods.Items[0].Spec.Containers {
		containerList = append(containerList, container.Name)
	}
	return containerList
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

type Alert struct {
	AlertName string
	Severity  string
	StartsAt  string
	PodName   string
	Summary   string
}

type origAlert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    string            `json:"startsAt"`
	EndsAt      string            `json:"endsAt"`
}

func (kc *K8sClient) GetAlerts() []Alert {
	//execute command to get alerts -  kubectl exec -it -n fed-prometheus alertmanager-prometheus-alerts-0 -- sh -c "amtool -o json alert query -a --alertmanager.url http://localhost:9093"
	alertList := []Alert{}

	command := "sh -c \"amtool -o json alert query -a --alertmanager.url http://localhost:9093\""
	stdout, stderr, err := kc.ExecuteRemoteCommand("fed-prometheus", "alertmanager-prometheus-alerts-0", "alertmanager", command)

	if err != nil {
		fmt.Println(err)
		fmt.Println(stderr)
	}

	origAlerts := []origAlert{}
	// Unmarshal the json output
	err = json.Unmarshal([]byte(stdout), &origAlerts)
	if err != nil {
		fmt.Println(err)
	}

	for _, alert := range origAlerts {
		alertList = append(alertList, Alert{
			AlertName: alert.Labels["alertname"],
			Severity:  alert.Labels["severity"],
			StartsAt:  alert.StartsAt,
			PodName:   alert.Labels["pod"],
			Summary:   alert.Annotations["summary"],
		})
	}
	return alertList

}

func (kc *K8sClient) ExecuteRemoteCommand(namespace, pod, container, command string) (string, string, error) {
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	request := kc.Client.CoreV1().RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(pod).
		SubResource("exec").
		Param("container", container).
		VersionedParams(&v1.PodExecOptions{
			Command: []string{"/bin/sh", "-c", command},
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		}, scheme.ParameterCodec)
	exec, _ := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	_ = exec.Stream(remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})

	return buf.String(), errBuf.String(), nil
}

type RedisDbSizeInfo struct {
	PodName string
	Output  string
}

func (kc *K8sClient) GetRedisDbSize() []RedisDbSizeInfo {
	redis_namespace := "fed-redis-cluster"
	redis_container := "redis-node"

	returnSize := []RedisDbSizeInfo{}

	//get the list of pods from the redis namespace
	pods, err := kc.Client.CoreV1().Pods(redis_namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	for _, pod := range pods.Items {
		//execute command to get the redis db size
		command := fmt.Sprintf("redis-cli --cluster call --cluster-only-masters redis-cluster.%s.svc.cluster.local:6379 dbsize", redis_namespace)
		stdout, stderr, err := kc.ExecuteRemoteCommand(redis_namespace, pod.Name, redis_container, command)
		if err != nil {
			fmt.Println(err)
			fmt.Println(stderr)
		}
		returnSize = append(returnSize, RedisDbSizeInfo{
			PodName: pod.Name,
			Output:  stdout,
		})
	}
	return returnSize
}
func (kc *K8sClient) FlushRedisData() error {
	redis_namespace := "fed-redis-cluster"
	redis_container := "redis-node"

	returnSize := []RedisDbSizeInfo{}

	//get the list of pods from the redis namespace
	pods, err := kc.Client.CoreV1().Pods(redis_namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	for _, pod := range pods.Items {
		//execute command to flush redis data
		command := fmt.Sprintf("redis-cli --cluster call --cluster-only-masters redis-cluster.%s.svc.cluster.local:6379 flushall", redis_namespace)
		stdout, stderr, err := kc.ExecuteRemoteCommand(redis_namespace, pod.Name, redis_container, command)
		if err != nil {
			fmt.Println(err)
			fmt.Println(stderr)
			return err
		}
		returnSize = append(returnSize, RedisDbSizeInfo{
			PodName: pod.Name,
			Output:  stdout,
		})
	}
	return nil
}

// Node represents an individual Redis node in the cluster
type Node struct {
	ID         string   `json:"id"`
	IP         string   `json:"ip"`
	PodName    string   `json:"podName"`
	Port       string   `json:"port"`
	Role       string   `json:"role"`
	Slots      []string `json:"slots,omitempty"`      // Omit if empty
	PrimaryRef string   `json:"primaryRef,omitempty"` // Omit if empty
	Zone       string   `json:"zone"`
}

// Cluster represents the cluster information
type Cluster struct {
	LabelSelectorPath          string         `json:"labelSelectorPath"`
	MaxReplicationFactor       int            `json:"maxReplicationFactor"`
	MinReplicationFactor       int            `json:"minReplicationFactor"`
	Nodes                      []Node         `json:"nodes"`
	NumberOfPods               int            `json:"numberOfPods"`
	NumberOfPodsReady          int            `json:"numberOfPodsReady"`
	NumberOfPrimaries          int            `json:"numberOfPrimaries"`
	NumberOfPrimariesReady     int            `json:"numberOfPrimariesReady"`
	NumberOfRedisNodesRunning  int            `json:"numberOfRedisNodesRunning"`
	NumberOfReplicasPerPrimary map[string]int `json:"numberOfReplicasPerPrimary"`
	Status                     string         `json:"status"`
}

// Condition represents a condition for the cluster status
type Condition struct {
	LastProbeTime      time.Time `json:"lastProbeTime"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Message            string    `json:"message"`
	Reason             string    `json:"reason"`
	Status             string    `json:"status"`
	Type               string    `json:"type"`
}

// ClusterStatus represents the overall cluster status including conditions
type ClusterStatus struct {
	Cluster    Cluster     `json:"cluster"`
	Conditions []Condition `json:"conditions"`
	StartTime  time.Time   `json:"startTime"`
}
type RedisStatus struct {
	PrimariesConfigured   int
	ReplicasConfigured    int
	PodStatus             bool
	ClusterState          bool
	ClusterSlotsOk        int
	ClusterKnownNodes     int
	ClusterSize           int
	ClusterSlotsPfail     int
	ClusterSlotsFail      int
	NumberActiveZones     int
	RedisNodeDetails      []Node
	NumberZonesPrimaries  int
	NumberPrimariesInZone int
	PodDetails            map[string]missingDetails
}

type missingDetails struct {
	Worker string
	CPU    string
	Memory string
}

func (kc *K8sClient) GetRedisStatus() RedisStatus {
	redis_namespace := "fed-redis-cluster"
	customResourceName := "node-for-redis"
	customResource, err := kc.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    "db.ibm.com",
		Version:  "v1alpha1",
		Resource: "redisclusters",
	}).Namespace(redis_namespace).Get(context.Background(), customResourceName, metav1.GetOptions{})

	if err != nil {
		fmt.Println(err)
		return RedisStatus{}
	}

	//get the status of the custom resource

	// Assuming the custom resource has a status field
	status, found, err := unstructured.NestedMap(customResource.Object, "status")
	if err != nil || !found {
		log.Fatalf("Error fetching status field: %v", err)
	}

	var clusterStatus ClusterStatus
	temp, err := json.Marshal(status)

	err = json.Unmarshal(temp, &clusterStatus)
	if err != nil {
		log.Fatalf("Error unmarshalling status field: %v", err)
	}

	return RedisStatus{
		PrimariesConfigured: clusterStatus.Cluster.NumberOfPrimaries,
		ReplicasConfigured:  clusterStatus.Cluster.MaxReplicationFactor,
		PodStatus:           clusterStatus.Cluster.NumberOfPods == clusterStatus.Cluster.NumberOfPodsReady,
		ClusterState:        clusterStatus.Cluster.Status == "OK",
		ClusterSlotsOk:      len(clusterStatus.Cluster.Nodes),
		ClusterKnownNodes:   len(clusterStatus.Cluster.Nodes),
		ClusterSize:         clusterStatus.Cluster.NumberOfPods,
		ClusterSlotsPfail:   0,
		ClusterSlotsFail:    0,
		RedisNodeDetails:    clusterStatus.Cluster.Nodes,
		NumberActiveZones: func() int {
			zoneMap := make(map[string]bool)
			for _, node := range clusterStatus.Cluster.Nodes {
				zoneMap[node.Zone] = true
			}
			return len(zoneMap)
		}(),
		NumberZonesPrimaries:  1,
		NumberPrimariesInZone: clusterStatus.Cluster.NumberOfPrimaries,
		PodDetails: func(nodeList []Node) map[string]missingDetails {
			podDetails := make(map[string]missingDetails)

			for _, node := range nodeList {
				var nodeDetails missingDetails
				pod, err := kc.Client.CoreV1().Pods(redis_namespace).Get(context.Background(), node.PodName, metav1.GetOptions{})
				if err != nil {
					fmt.Println(err)
				}
				nodeDetails.Worker = pod.Spec.NodeName
				nodeDetails.CPU = pod.Spec.Containers[0].Resources.Requests.Cpu().String()
				nodeDetails.Memory = pod.Spec.Containers[0].Resources.Requests.Memory().String()
				podDetails[node.PodName] = nodeDetails
			}
			return podDetails

		}(clusterStatus.Cluster.Nodes),
	}

}

// cmd = 'kubectl -n {} exec -it {} -c {} bash -- curl http://127.0.0.1:{}/tenv/eTrace/enable?filter=all\&level=DEBUG_{}'.format(namespace, pod_name, pod_config['container'], pod_config['port'], debug_level)
func (kc *K8sClient) SetDebugLevel(namespace, pod, container, debugLevel string) bool {
	port := "9090"
	//TODO: Port is hardcoded here. It should be fetched from the local config based on the service name
	//TODO: Prepare the local config file to fetch the port based on the service name
	command := fmt.Sprintf("curl http://127.0.0.1:%s/tenv/eTrace/enable?filter=all\\&level=%s", port, debugLevel)
	fmt.Println(command)
	stdout, stderr, err := kc.ExecuteRemoteCommand(namespace, pod, container, command)
	if err != nil {
		fmt.Println(err)
		fmt.Println(stderr)
		return false
	}

	fmt.Println(stdout)
	return true
}

func (kc *K8sClient) GetKargoServiceIP() (string, error) {
	service, err := kc.Client.CoreV1().Services("fed-paas-helpers").Get(context.Background(), "kargo", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if len(service.Status.LoadBalancer.Ingress) == 0 {
		return "", fmt.Errorf("no LoadBalancer Ingress found for kargo service")
	}

	return service.Status.LoadBalancer.Ingress[0].IP, nil
}
