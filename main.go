package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/nicolasmmb/kube-port-forward/config"
	"github.com/nicolasmmb/kube-port-forward/models"
	"github.com/nicolasmmb/kube-port-forward/utils/check"
	"github.com/nicolasmmb/kube-port-forward/utils/prompt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	pfclt "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func main() {

	cfg := models.BaseConfig{}
	cfg.LoadConfig()

	kubeConfig, kubeClient := config.CreateKubernetesSession(cfg.Kubernetes.ConfigFile.Directory)

	// Namespaces
	choicesNamespace := []string{}
	for _, terminal := range cfg.Kubernetes.Namespaces {
		choicesNamespace = append(choicesNamespace, terminal.Name)
	}
	_, selectedNamespace, err := prompt.Select("Select the Namespace", choicesNamespace)
	check.Error(err, true, true)

	// Pods
	pods, err := kubeClient.CoreV1().Pods(selectedNamespace).List(context.Background(), v1.ListOptions{})
	check.Error(err, true, true)

	podsChoices := []string{}
	for _, pod := range pods.Items {
		podsChoices = append(podsChoices, pod.Name)
	}
	selectedPodIndex, _, err := prompt.Select("Select the Pod", podsChoices)
	check.Error(err, true, true)

	actualPod := pods.Items[selectedPodIndex]

	// Ports
	choicesPorts := []string{}
	for _, container := range actualPod.Spec.Containers {
		for _, port := range container.Ports {
			choicesPorts = append(choicesPorts, strconv.Itoa(int(port.ContainerPort)))
		}
	}

	_, selectedPort, err := prompt.Select("Select the Port", choicesPorts)
	check.Error(err, true, true)

	prompt.Confirm("Start PortForward", true, "y")

	// PortForward
	pf := config.PortFowardConfig(
		"0.0.0.0",
		selectedPort,
		kubeConfig,
		kubeClient,
	)

	kubeRequest := KubeRequestPortForward(
		kubeClient,
		actualPod.Name,
		actualPod.Namespace,
	)

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

func KubeRequestPortForward(
	client *kubernetes.Clientset,
	podName string,
	podNamespace string,
) *url.URL {
	return client.RESTClient().Post().
		Prefix("api/v1").
		Resource("pods").
		SubResource("portforward").
		Namespace(podNamespace).
		Name(podName).URL()
}
