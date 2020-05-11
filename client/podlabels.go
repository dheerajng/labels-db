package client

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

// K8sClient holds kubeconfig and clientset
type K8sClient struct {
	config    *rest.Config
	clientset *kubernetes.Clientset
}

// NewK8sClient creates kubernetes client which will be used to retrieve labels' info from API server
func NewK8sClient(kubeConfig []byte, contextName string) (*K8sClient, error) {
	client := K8sClient{}
	config, err := getConfig(kubeConfig, contextName)
	if err != nil {
		logrus.Errorf("Could not get config properly. %s", err)
		return nil, err
	}
	client.config = config

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Errorf("Could not create clientset. %s", err)
		return nil, err
	}
	client.clientset = clientset
	return &client, nil
}

func getConfig(kubeconfig []byte, contextName string) (*rest.Config, error) {
	if len(kubeconfig) > 0 {
		logrus.Debugf("kubeconfig is provided")
		cfg, err := clientcmd.Load(kubeconfig)
		if err != nil {
			return nil, err
		}
		if contextName != "" {
			cfg.CurrentContext = contextName
		}
		return clientcmd.NewDefaultClientConfig(*cfg, &clientcmd.ConfigOverrides{}).ClientConfig()
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Errorf("Could not get config. %s", err)
		return nil, err
	}
	return config, nil
}

// GetPodsLabels returns labels of a pods
func (lClient *K8sClient) GetPodsLabels() error {
	for {
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		t1 := time.Now()
		pods, err := lClient.clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		/*
			pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
				FieldSelector: "Status.Phase: Running"},
			)
		*/
		if err != nil {
			logrus.Errorf("Could not retrieve pods' info")
			return err
		}
		t2 := time.Now()
		diff := t2.Sub(t1)
		fmt.Println("Time taken to retrieve pods in all namespaces: ", len(pods.Items), diff)
		for i, pod := range pods.Items {
			fmt.Printf("Pod %d: ", i)
			if pod.Status.Phase != api.PodRunning {
				fmt.Printf("Pod %s is not running.\n", pod.GetName())
				continue
			}
			labels := pod.ObjectMeta.GetLabels()
			fmt.Printf("Podname: %s, PodIP: %s, PodLabels: %v\n", pod.GetName(), pod.Status.PodIP, labels)
		}

		// Examples for error handling:
		// - Use helper functions e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		/*
			_, err = clientset.CoreV1().Pods("default").Get("example-xxxxx", metav1.GetOptions{})
			if errors.IsNotFound(err) {
				fmt.Printf("Pod example-xxxxx not found in default namespace\n")
			} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
				fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
			} else if err != nil {
				fmt.Printf("Panic... \n")
				panic(err.Error())
			} else {
				fmt.Printf("Found example-xxxxx pod in default namespace\n")
			}
		*/
		time.Sleep(30 * time.Second)
	}

}
