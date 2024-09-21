package testsuite

import (
	"context"

	"healthctl/pkg/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CheckPAAS(clientset *kubernetes.Clientset) []models.ResourceCheck {
	checks := []models.ResourceCheck{
		CheckGrafana(clientset),
		CheckKibana(clientset),
		CheckPrometheus(clientset),
		CheckDbEtcd(clientset),
		CheckIstio(clientset),
		CheckKubeProm(clientset),
		CheckRedisOperator(clientset),
		CheckRedisCluster(clientset),
		CheckJaeger(clientset),
		CheckElastic(clientset),
		CheckElastAlert(clientset),
		CheckAlerta(clientset),
		CheckKiali(clientset),
	}
	return checks
}

// Check functions
func CheckGrafana(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if Grafana pod is running in fed-grafana namespace
	pods, err := clientset.CoreV1().Pods("fed-grafana").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Grafana", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"Grafana", "No Grafana pods found", false}
	}

	// Check if Grafana service is up
	services, err := clientset.CoreV1().Services("fed-grafana").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Grafana", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"Grafana", "No Grafana services found", false}
	}

	return models.ResourceCheck{"Grafana", "Grafana is Up", true}
}

func CheckKibana(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if Kibana pod is running in fed-kibana namespace
	pods, err := clientset.CoreV1().Pods("fed-kibana").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Kibana", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"Kibana", "No Kibana pods found", false}
	}

	// Check if Kibana service is up
	services, err := clientset.CoreV1().Services("fed-kibana").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Kibana", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"Kibana", "No Kibana services found", false}
	}

	return models.ResourceCheck{"Kibana", "Kibana is Up", true}
}

func CheckPrometheus(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if Prometheus pod is running in fed-prometheus namespace
	pods, err := clientset.CoreV1().Pods("fed-prometheus").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Prometheus", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"Prometheus", "No Prometheus pods found", false}
	}

	// Check if Prometheus service is up
	services, err := clientset.CoreV1().Services("fed-prometheus").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Prometheus", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"Prometheus", "No Prometheus services found", false}
	}

	return models.ResourceCheck{"Prometheus", "Prometheus is Up", true}
}

func CheckDbEtcd(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if etcd pod is running in fed-etcd namespace
	pods, err := clientset.CoreV1().Pods("fed-etcd").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Etcd", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"Etcd", "No Etcd pods found", false}
	}

	// Check if etcd service is up
	services, err := clientset.CoreV1().Services("fed-etcd").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Etcd", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"Etcd", "No Etcd services found", false}
	}

	return models.ResourceCheck{"Etcd", "Etcd is Up", true}
}

func CheckIstio(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if Istio pod is running in fed-istio-system namespace
	pods, err := clientset.CoreV1().Pods("fed-istio-system").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Istio", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"Istio", "No Istio pods found", false}
	}

	// Check if Istio service is up
	services, err := clientset.CoreV1().Services("fed-istio-system").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Istio", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"Istio", "No Istio services found", false}
	}

	return models.ResourceCheck{"Istio", "Istio is Up", true}
}

func CheckKubeProm(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if KubeProm pod is running in fed-kube-prom namespace
	pods, err := clientset.CoreV1().Pods("fed-kube-prom").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"KubeProm", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"KubeProm", "No KubeProm pods found", false}
	}

	// Check if KubeProm service is up
	services, err := clientset.CoreV1().Services("fed-kube-prom").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"KubeProm", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"KubeProm", "No KubeProm services found", false}
	}

	return models.ResourceCheck{"KubeProm", "KubeProm is Up", true}
}

func CheckRedisOperator(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if RedisOperator pod is running in fed-redis-operator namespace
	pods, err := clientset.CoreV1().Pods("fed-redis-operator").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"RedisOperator", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"RedisOperator", "No RedisOperator pods found", false}
	}

	// Check if RedisOperator service is up
	services, err := clientset.CoreV1().Services("fed-redis-operator").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"RedisOperator", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"RedisOperator", "No RedisOperator services found", false}
	}

	return models.ResourceCheck{"RedisOperator", "RedisOperator is Up", true}
}

func CheckRedisCluster(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if RedisCluster pod is running in fed-redis-cluster namespace
	pods, err := clientset.CoreV1().Pods("fed-redis-cluster").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"RedisCluster", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"RedisCluster", "No RedisCluster pods found", false}
	}

	// Check if RedisCluster service is up
	services, err := clientset.CoreV1().Services("fed-redis-cluster").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"RedisCluster", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"RedisCluster", "No RedisCluster services found", false}
	}

	return models.ResourceCheck{"RedisCluster", "RedisCluster is Up", true}
}

func CheckJaeger(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if Yaeger pod is running in fed-yaeger namespace
	pods, err := clientset.CoreV1().Pods("fed-yaeger").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Yaeger", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"Yaeger", "No Yaeger pods found", false}
	}

	// Check if Yaeger service is up
	services, err := clientset.CoreV1().Services("fed-yaeger").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Yaeger", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"Yaeger", "No Yaeger services found", false}
	}

	return models.ResourceCheck{"Yaeger", "Yaeger is Up", true}
}

func CheckElastic(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if Elastic pod is running in fed-elastic namespace
	pods, err := clientset.CoreV1().Pods("fed-elastic").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Elastic", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"Elastic", "No Elastic pods found", false}
	}

	// Check if Elastic service is up
	services, err := clientset.CoreV1().Services("fed-elastic").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Elastic", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"Elastic", "No Elastic services found", false}
	}

	return models.ResourceCheck{"Elastic", "Elastic is Up", true}
}

func CheckElastAlert(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if ElastAlert pod is running in fed-elastalert namespace
	pods, err := clientset.CoreV1().Pods("fed-elastalert").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"ElastAlert", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"ElastAlert", "No ElastAlert pods found", false}
	}

	// Check if ElastAlert service is up
	services, err := clientset.CoreV1().Services("fed-elastalert").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"ElastAlert", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"ElastAlert", "No ElastAlert services found", false}
	}

	return models.ResourceCheck{"ElastAlert", "ElastAlert is Up", true}
}

func CheckAlerta(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if Alerta pod is running in fed-alerta namespace
	pods, err := clientset.CoreV1().Pods("fed-alerta").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Alerta", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"Alerta", "No Alerta pods found", false}
	}

	// Check if Alerta service is up
	services, err := clientset.CoreV1().Services("fed-alerta").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Alerta", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"Alerta", "No Alerta services found", false}
	}

	return models.ResourceCheck{"Alerta", "Alerta is Up", true}
}

func CheckKiali(clientset *kubernetes.Clientset) models.ResourceCheck {

	// Check if Kiali pod is running in fed-kiali namespace
	pods, err := clientset.CoreV1().Pods("fed-kiali").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Kiali", "Error fetching pods", false}
	}

	if len(pods.Items) == 0 {
		return models.ResourceCheck{"Kiali", "No Kiali pods found", false}
	}

	// Check if Kiali service is up
	services, err := clientset.CoreV1().Services("fed-kiali").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return models.ResourceCheck{"Kiali", "Error fetching services", false}
	}

	if len(services.Items) == 0 {
		return models.ResourceCheck{"Kiali", "No Kiali services found", false}
	}

	return models.ResourceCheck{"Kiali", "Kiali is Up", true}
}
