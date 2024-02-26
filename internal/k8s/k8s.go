package k8s

import (
	"context"
	"flag"
	"log"
	"path/filepath"

	"github.com/michaeljsaenz/probe/internal/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func init() {
	getClientSet()
}

func getClientSet() {
	// https://github.com/kubernetes/client-go/blob/master/examples/out-of-cluster-client-configuration/main.go
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		//TODO raise this error to UI
		log.Fatal("kubeconfig error: ", err)
	}

	// Override the TLSClientConfig to skip certificate verification
	config.TLSClientConfig = rest.TLSClientConfig{Insecure: true}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		//TODO raise this error to UI
		log.Fatal("clientset error: ", err)
	}

	types.UpdateSharedContextK8s(clientset, "")

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
func GetPodsInNamespace(c *kubernetes.Clientset, namespace string) ([]string, error) {
	var podData []string

	pods, err := c.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		if err != nil {
			return []string{}, err
		}
	}
	for _, pod := range pods.Items {
		podData = append(podData, pod.Name)
	}

	return podData, nil

}
