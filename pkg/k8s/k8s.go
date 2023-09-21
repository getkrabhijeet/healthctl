package k8s
import (

	"flag"
	"path/filepath"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

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

func GetClusterhealth(client *kubernetes.Clientset) {
	// Get cluster version
	clusterVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Cluster version: %s\n", clusterVersion)

	// Get cluster nodes
	nodes, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, node := range nodes.Items {
		fmt.Printf("Node name: %s\n", node.Name)
	}

	// Get cluster namespaces
	namespaces, err := client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, namespace := range namespaces.Items {
		fmt.Printf("Namespace name: %s\n", namespace.Name)
	}




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


	//  Print all resources in the cluster
	// k8shelpers.PrintPods(client)
	// k8shelpers.PrintDeployments(client)
	// k8shelpers.PrintStatefulSets(client)
	// k8shelpers.PrintDaemonSets(client)
	// k8shelpers.PrintServices(client)
	// k8shelpers.PrintIngresses(client)
	// k8shelpers.PrintConfigMaps(client)
	// k8shelpers.PrintSecrets(client)
	// k8shelpers.PrintPVCs(client)
	// k8shelpers.PrintPVs(client)
	// k8shelpers.PrintJobs(client)
	// k8shelpers.PrintCronJobs(client)
	// k8shelpers.PrintClusterRoles(client)
	// k8shelpers.PrintClusterRoleBindings(client)
	// k8shelpers.PrintRoles(client)
	// k8shelpers.PrintRoleBindings(client)
	// k8shelpers.PrintCRDs(client)
	// k8shelpers.PrintCRs(client)
	// k8shelpers.PrintStorageClasses(client)
	// k8shelpers.PrintNodes(client)
	// k8shelpers.PrintNamespaces(client)
	// k8shelpers.PrintEvents(client)
	// k8shelpers.PrintEndpoints(client)
	// k8shelpers.PrintLimitRanges(client)
	// k8shelpers.PrintResourceQuotas(client)
	// k8shelpers.PrintPodDisruptionBudgets(client)
	// k8shelpers.PrintPodSecurityPolicies(client)
	// k8shelpers.PrintPriorityClasses(client)
	// k8shelpers.PrintServiceAccounts(client)
	// k8shelpers.PrintMutatingWebhookConfigurations(client)
	// k8shelpers.PrintValidatingWebhookConfigurations(client)
	// k8shelpers.PrintPodTemplates(client)
	// k8shelpers.PrintReplicaSets(client)
	// k8shelpers.PrintReplicationControllers(client)
	// k8shelpers.PrintNetworkPolicies(client)