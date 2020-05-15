package client

import (
	"fmt"
	rc "labels-db/redisclient"
	"time"

	"github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

// PodDetails stores info for pod IP
type PodDetails struct {
	IP        string            `json:"ip"`
	PodName   string            `json:"podname"`
	Service   string            `json:"service"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
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
	fmt.Print("\n\n########################################\n\n")
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
	return nil
}

// GetSvcDetails retrieves list of all services
func (lClient *K8sClient) GetSvcDetails() error {
	t1 := time.Now()
	services, err := lClient.clientset.CoreV1().Services("").List(metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Could not retrieve Services")
		return err
	}
	t2 := time.Now()
	diff := t2.Sub(t1)
	fmt.Println("Time taken to retrieve services in all namespaces: ", len(services.Items), diff)
	for i, svc := range services.Items {
		fmt.Printf("Service %d: ", i)
		labels := svc.ObjectMeta.GetLabels()
		fmt.Printf("ServiceName: %s, Namespace: %s, ServiceLabels: %v, Selector: %v\n", svc.GetName(), svc.GetNamespace(), labels, svc.Spec.Selector)
		// Get pods of this service
		podinfo, err := lClient.GetPodsDetails(svc.GetName(), svc.GetNamespace(), svc.Spec.Selector)
		if err != nil {
			logrus.Debugf("Could not list pods for service %s\n", svc.GetName())
		}
		fmt.Println(podinfo)
		fmt.Println("------------")
	}
	t3 := time.Now()
	diff = t3.Sub(t1)
	fmt.Println("Time taken to list all services and pods: ", diff)
	fmt.Println("#######################")
	return nil
}

// GetPodsDetails retrieves pods of a service
func (lClient *K8sClient) GetPodsDetails(service string, namespace string, selector map[string]string) ([]PodDetails, error) {
	var podsInfo []PodDetails
	logrus.Debugf("Trying to retrieve pods for service %s.%s with labels %v", service, namespace, selector)
	if len(selector) == 0 {
		logrus.Debugf("Not selecting any pod for service %s due to Empty Selector", service)
		return nil, nil
	}
	set := labels.Set(selector)
	pods, err := lClient.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: set.String()})
	if err != nil {
		logrus.Errorf("Could not retrieve pods of service %s", service)
		return nil, err
	}
	for i, pod := range pods.Items {
		fmt.Printf("\tPod %d: ", i)
		if pod.Status.Phase != api.PodRunning {
			fmt.Printf("Pod %s is not running.\n", pod.GetName())
			continue
		}
		labels := pod.ObjectMeta.GetLabels()
		pi := PodDetails{IP: pod.Status.PodIP, PodName: pod.GetName(), Service: service, Namespace: namespace, Labels: labels}
		podsInfo = append(podsInfo, pi)
		//fmt.Printf("Podname: %s, PodIP: %s, PodLabels: %v\n", pod.GetName(), pod.Status.PodIP, labels)
		// Insert in DB
		_ = insertInDB(pi)
	}
	return podsInfo, nil
}

// GetDeploymentDetails retrieves deployments in namespace
func (lClient *K8sClient) GetDeploymentDetails(namespace string, selector map[string]string) error {
	logrus.Debugf("Trying to retrieve deployments in namespace %s with labels %v", namespace, selector)
	set := labels.Set(selector)
	dpls, err := lClient.clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{LabelSelector: set.String()})
	if err != nil {
		logrus.Errorf("Could not retrieve deployments with selector %v", selector)
		return err
	}
	for i, dpl := range dpls.Items {
		fmt.Printf("\n\t")
		fmt.Printf("%d) Deployment: %s, Labels: %v, Spec: %v", i, dpl.GetName(), dpl.GetLabels(), dpl.Spec.Selector)
	}
	fmt.Println("---")
	return nil
}

// CreateLablesDB will generate DB entries for each running pod IP
func (lClient *K8sClient) CreateLablesDB() error {
	if err := lClient.GetSvcDetails(); err != nil {
		logrus.Errorf("Could not create Labels Database. %s", err.Error())
		return err
	}
	return nil
}

func insertInDB(podinfo PodDetails) error {
	key := podinfo.IP
	value := rc.PodDBValue{PodName: podinfo.PodName, Service: podinfo.Service, Namespace: podinfo.Namespace, Labels: podinfo.Labels}
	err := rc.SetStruct(key, value)
	if err != nil {
		logrus.Errorf("Could not insert %v in DB. Err: %s", podinfo, err.Error())
	}
	return nil
}

// GetOneFromDB will retrieve the details for single IP
func GetOneFromDB(IP string) (*PodDetails, error) {
	podinfo := PodDetails{}
	value, err := rc.GetStruct(IP)
	if err != nil {
		logrus.Errorf("Could not retrieve details from DB for IP: %s", IP)
		return nil, err
	}
	podinfo = PodDetails{IP: string(IP), PodName: value.PodName, Service: value.Service, Namespace: value.Namespace, Labels: value.Labels}
	return &podinfo, nil
}

// GetMultiFromDB will retrieve the detais for multiple IPs
func GetMultiFromDB(IPs []string) ([]PodDetails, error) {
	var values []rc.PodDBValue
	var podsInfo []PodDetails
	values, err := rc.GetMultiStruct(IPs)
	if err != nil {
		logrus.Errorf("Could not retrieve details from DB for %v", IPs)
		return nil, err
	}
	for i, value := range values {
		podsInfo = append(podsInfo, PodDetails{IP: string(IPs[i]), PodName: value.PodName, Service: value.Service, Namespace: value.Namespace, Labels: value.Labels})
	}
	return podsInfo, nil
}
