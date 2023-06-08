package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/nicolasmmb/kube-port-forward/models"
	"github.com/nicolasmmb/kube-port-forward/utils/check"
	"github.com/nicolasmmb/kube-port-forward/utils/prompt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	pfclt "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/cmd/portforward"
)

func main() {

	cfg := models.BaseConfig{}
	cfg.LoadConfig()
	cfg.PrintConfig()

	kubeConfig, kubeClient, err := ConfigureKubernetesSession(cfg.Kubernetes.ConfigFile.Directory)
	check.Error(err, true, true)

	choicesNamespace := []string{}
	for _, terminal := range cfg.Kubernetes.Namespaces {
		choicesNamespace = append(choicesNamespace, terminal.Name)
	}

	_, selectedNamespace, err := prompt.Select("Select the Namespace", choicesNamespace)
	check.Error(err, true, true)

	pods, err := kubeClient.CoreV1().Pods(selectedNamespace).List(context.Background(), v1.ListOptions{})
	check.Error(err, true, true)

	podsChoices := []string{}

	for _, pod := range pods.Items {
		podsChoices = append(podsChoices, pod.Name)
	}

	selectedPodIndex, _, err := prompt.Select("Select the Pod", podsChoices)
	check.Error(err, true, true)

	actualPod := pods.Items[selectedPodIndex]

	choicesPorts := []string{}
	for _, container := range actualPod.Spec.Containers {
		for _, port := range container.Ports {

			choicesPorts = append(choicesPorts, strconv.Itoa(int(port.ContainerPort)))
		}
	}

	_, selectedPort, err := prompt.Select("Select the Port", choicesPorts)
	check.Error(err, true, true)

	prompt.Confirm("Start PortForward", true, "y")

	pf := portforward.PortForwardOptions{}

	pf.Address = []string{"0.0.0.0"}
	pf.Ports = []string{selectedPort + ":" + selectedPort}
	pf.PodName = actualPod.Name
	pf.Namespace = actualPod.Namespace
	pf.ReadyChannel = make(chan struct{}, 1)
	pf.StopChannel = make(chan struct{}, 1)
	pf.PodClient = kubeClient.CoreV1()

	kubeRequest := kubeClient.RESTClient().Post().
		Prefix("api/v1").
		Resource("pods").
		SubResource("portforward").
		Namespace(actualPod.Namespace).
		Name(actualPod.Name).
		URL()

	transport, upgrader, err := spdy.RoundTripperFor(kubeConfig)
	check.Error(err, true, true)

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{
			Transport: transport,
		}, http.MethodPost, &url.URL{
			Scheme: "https",
			Path:   kubeRequest.Path,
			Host:   kubeRequest.Host,
		},
	)

	fw, err := pfclt.New(dialer, pf.Ports, pf.StopChannel, pf.ReadyChannel, os.Stdout, os.Stderr)
	check.Error(err, true, true)

	err = fw.ForwardPorts()
	check.Error(err, true, true)

}

func ConfigureKubernetesSession(kubeconfigPath string) (*rest.Config, *kubernetes.Clientset, error) {

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	check.Error(err, true, true)

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	check.Error(err, true, true)

	return kubeConfig, kubeClient, err
}
