package main

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/nicolasmmb/kube-port-forward/models"
	"github.com/nicolasmmb/kube-port-forward/utils/check"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	pfclt "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/cmd/portforward"
)

func main() {

	path, err := os.UserHomeDir()
	check.Error(err, true, true)
	path += "/.kube/config"

	cfg := models.BaseConfig{}
	cfg.LoadConfig()
	cfg.PrintConfig()

	kubeConfig, kubeClient, err := ConfigureKubernetesSession(path)
	check.Error(err, true, true)

	choicesNamespace := []string{}
	for _, terminal := range cfg.Kubernetes.Namespaces {
		choicesNamespace = append(choicesNamespace, terminal.Name)
	}

	promptNamespace := promptui.Select{
		Label: "Selecione o namespace",
		Items: choicesNamespace,
		Size:  10,
	}

	_, resultNamespace, err := promptNamespace.Run()
	check.Error(err, true, true)

	pods, err := kubeClient.CoreV1().Pods(resultNamespace).List(context.Background(), v1.ListOptions{})
	check.Error(err, true, true)

	podsChoices := []string{}

	for _, pod := range pods.Items {
		podsChoices = append(podsChoices, pod.Name)
	}

	promptPod := promptui.Select{
		Label: "Selecione o pod",
		Items: podsChoices,
		Size:  10,
	}

	podSelectedIndex, _, err := promptPod.Run()
	check.Error(err, true, true)

	actualPod := pods.Items[podSelectedIndex]

	allPorts := []int32{8000}
	for _, container := range actualPod.Spec.Containers {
		for _, port := range container.Ports {
			allPorts = append(allPorts, port.ContainerPort)
		}
	}

	// if len(allPorts) == 0 {
	// 	fmt.Println("Não há portas disponíveis")
	// 	os.Exit(1)
	// }

	promptPort := promptui.Select{
		Label: "Selecione a porta",
		Items: allPorts,
		Size:  10,
	}

	_, selectedPort, err := promptPort.Run()
	check.Error(err, true, true)

	promptPortFoward := promptui.Prompt{
		Label:     "Deseja iniciar o port foward",
		IsConfirm: true,
		Default:   "y",
	}

	_, err = promptPortFoward.Run()
	check.Error(err, true, true)

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
