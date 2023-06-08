package config

import (
	"github.com/nicolasmmb/kube-port-forward/utils/check"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ptfw "k8s.io/kubectl/pkg/cmd/portforward"
)

func CreateKubernetesSession(kubeconfigPath string) (*rest.Config, *kubernetes.Clientset) {

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	check.Error(err, true, true)

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	check.Error(err, true, true)

	return kubeConfig, kubeClient
}

func PortFowardConfig(
	address string,
	port string,
	kubeConfig *rest.Config,
	kubeClient *kubernetes.Clientset) *ptfw.PortForwardOptions {

	port = port + ":" + port

	pf := ptfw.PortForwardOptions{}
	pf.Address = []string{address}
	pf.Ports = []string{port}
	pf.ReadyChannel = make(chan struct{}, 1)
	pf.StopChannel = make(chan struct{}, 1)
	pf.PodClient = kubeClient.CoreV1()

	return &pf
}
