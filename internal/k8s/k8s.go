package k8s

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/michaeljsaenz/probe/internal/types"
	"github.com/michaeljsaenz/probe/internal/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
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

	types.UpdateSharedContextK8s(clientset, config, namespace)

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

	podContainers := make(map[string][]int32)

	for _, container := range pod.Spec.Containers {
		podContainers[container.Name] = []int32{}
		for _, port := range container.Ports {

			podContainers[container.Name] = append(podContainers[container.Name], port.ContainerPort)
		}

	}

	return types.PodDetail{
		PodName:       podName,
		PodNamespace:  podNamespace,
		PodStatus:     string(pod.Status.Phase),
		PodAge:        podAge,
		PodNode:       pod.Spec.NodeName,
		PodContainers: podContainers,
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

func PortForward(clientset *kubernetes.Clientset, config *rest.Config,
	namespace string, podName string, containerPort string) (URL, podPort string, err error) {

	var startinglocalPort int = 9000
	localPort, err := utils.FindLocalPort(startinglocalPort)
	if err != nil {
		log.Printf("error:%v", err)
		return "", "", err
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	// create roundtripper/upgrader for spdy
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return "", "", fmt.Errorf("failed to create round tripper: %v", err)
	}

	// channel for status and outputs
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})
	out := new(strings.Builder)
	errOut := new(strings.Builder)

	// create the port forwarder
	pf, err := portforward.New(
		spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, req.URL()),
		[]string{fmt.Sprintf("%s:%s", localPort, containerPort)},
		stopChan,
		readyChan,
		out,
		errOut,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to create port forwarder: %v", err)
	}

	// start port forward
	go func() {
		if err := pf.ForwardPorts(); err != nil {
			log.Printf("port forward failed: %v", err)
		}
	}()

	// close port forward
	go func() {
		time.Sleep(300 * time.Second)
		close(stopChan)
		log.Printf("Forwarding from http://127.0.0.1:%s -> %s (closed)", localPort, containerPort)
	}()

	// wait until port forward is ready
	select {
	case <-readyChan:
		log.Printf("Forwarding from http://127.0.0.1:%s -> %s", localPort, containerPort)
		URL := fmt.Sprintf("http://127.0.0.1:%s", localPort)
		return URL, containerPort, nil
	case <-time.After(10 * time.Second):
		return "", "", fmt.Errorf("timeout waiting for port forward to be ready")
	}
}

func GetContainerLog(clientset *kubernetes.Clientset, podName, containerName, namespace string) (string, error) {
	logOptions := &corev1.PodLogOptions{Container: containerName}
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		log.Printf("error:%v", err)
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		log.Printf("error:%v", err)
		return "", err
	}
	return buf.String(), nil
}

func ContainerExec(clientset *kubernetes.Clientset, config *rest.Config,
	podName, containerName, namespace string, command string) (string, error) {
	exeCommand := []string{"sh", "-c", command}
	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName).
		Param("stdout", "true").
		Param("stderr", "true").
		Param("stdin", "true").
		Param("tty", "false")

	for _, cmd := range exeCommand {
		req.Param("command", cmd)
	}

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("error creating SPDY executor: %v", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
		Tty:    false,
	})

	if err != nil {
		return "", fmt.Errorf("error executing command in container: %v", err)
	}

	var combinedOutput bytes.Buffer
	combinedOutput.Write(stdoutBuf.Bytes())
	combinedOutput.Write(stderrBuf.Bytes())

	return combinedOutput.String(), nil
}
