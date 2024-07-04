package k8s

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"time"

	"github.com/michaeljsaenz/probe/internal/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

func init() {
	initClientSet()
}

var (
	kubeconfig *string
)

func initClientSet() {
	// https://github.com/kubernetes/client-go/blob/master/examples/out-of-cluster-client-configuration/main.go
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
}

func GetClientSet() {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal("kubeconfig error: ", err)
	}

	// Override the TLSClientConfig to skip certificate verification
	config.TLSClientConfig = rest.TLSClientConfig{Insecure: true}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("clientset error: ", err)
	}

	var namespace string

	//retrieve k8s clientset/namespace from shared context
	customValues, ok := types.SharedContextK8s.Value(types.ContextKey).(types.CustomContextValuesK8s)
	if ok {
		namespace = customValues.Namespace
	}

	types.UpdateSharedContextK8s(clientset, namespace)

}

// refreshClientSet periodically refreshes the Kubernetes clientset
func RefreshClientSet() {
	for {
		time.Sleep(15 * time.Second)
		GetClientSet()

	}

}

// get namespaces
func GetNamespaces(c *kubernetes.Clientset) (namespaces []string, err error) {
	// retrieve the list of namespaces
	namespaceList, err := c.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})

	if err != nil {
		return []string{}, err
	} else {
		for _, namespace := range namespaceList.Items {
			namespaces = append(namespaces, namespace.Name)
		}
	}
	return namespaces, err

}

// get pod names with provided namespace
func GetPodsInNamespace(c *kubernetes.Clientset, namespace string) ([]types.K8sPod, error) {
	var K8sPods []types.K8sPod

	if namespace == "all namespaces" {
		namespace = ""
	}

	pods, err := c.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})

	if err != nil {
		return []types.K8sPod{}, err
	}
	for _, pod := range pods.Items {
		k8sPod := types.K8sPod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
		}
		K8sPods = append(K8sPods, k8sPod)
	}

	return K8sPods, nil

}

// get nodes
func GetNodes(c *kubernetes.Clientset) ([]types.K8sNode, types.K8sNodesDetail, error) {
	var K8sNodes []types.K8sNode
	var nodesDetail types.K8sNodesDetail

	nodes, err := c.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	nodesDetail.TotalCount = len(nodes.Items)

	if err != nil {
		return []types.K8sNode{}, types.K8sNodesDetail{}, err
	}
	for _, node := range nodes.Items {
		status := "Unknown"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			} else if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionFalse {
				status = "NotReady"
				break
			}
		}
		k8sNode := types.K8sNode{
			Name:   node.Name,
			Status: status,
		}
		K8sNodes = append(K8sNodes, k8sNode)
	}

	return K8sNodes, nodesDetail, nil

}

func GetPodDetail(c *kubernetes.Clientset, podNamespace, podName string) (types.PodDetail, error) {
	pod, err := c.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, v1.GetOptions{})

	if err != nil {
		log.Printf("failed to get pod detail: %v", err)
		return types.PodDetail{}, err

	}

	podCreationTime := pod.GetCreationTimestamp()
	age := time.Since(podCreationTime.Time).Round(time.Second)
	podAge := age.String()
	if int(math.Trunc(age.Hours())) >= 24 {
		ageInDays := int(math.Trunc(age.Hours())) / 24
		podAge = strconv.Itoa(ageInDays) + "d"
	}

	var containers []string
	for _, container := range pod.Spec.Containers {
		containers = append(containers, container.Name)
	}

	return types.PodDetail{
		PodStatus:     string(pod.Status.Phase),
		PodAge:        podAge,
		PodNode:       pod.Spec.NodeName,
		PodContainers: containers,
	}, nil
}

func GetPodYaml(c *kubernetes.Clientset, podNamespace string, podName string) (string, error) {
	pod, err := c.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, v1.GetOptions{})
	if err != nil {
		log.Printf("failed to get pod yaml: %v", err)
		return "", err
	}

	// clear unnecessary fields
	pod.ObjectMeta.ManagedFields = nil
	pod.ObjectMeta.GenerateName = ""
	pod.Status = corev1.PodStatus{}

	// serialize the Pod to YAML format
	codec := serializer.NewCodecFactory(scheme.Scheme).LegacyCodec(corev1.SchemeGroupVersion)
	marshaledYaml, err := runtime.Encode(codec, pod)
	if err != nil {
		return "", fmt.Errorf("error encoding YAML: %v", err)
	}

	// convert the marshaled YAML to a string
	yamlString, err := yaml.JSONToYAML(marshaledYaml)
	if err != nil {
		log.Printf("error converting YAML to string: %v", err)
		return "", err
	}

	return string(yamlString), nil

}
