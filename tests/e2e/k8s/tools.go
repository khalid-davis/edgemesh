package k8s

import (
	"encoding/json"
	"fmt"
	"github.com/kubeedge/kubeedge/tests/e2e/constants"
	"github.com/kubeedge/kubeedge/tests/e2e/utils"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
	"time"
)

var (
	defaultNamespace = "default"
)

// busybox
func generateBusybox(name string, labels, nodeSelector map[string]string) *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: labels,
		},
		Spec: v1.PodSpec{
			NodeSelector: nodeSelector,
			Containers: []v1.Container{
				{
					Name:            "busybox",
					Image:           "sequenceiq/busybox",
					ImagePullPolicy: "IfNotPresent",
					Args:            []string{"sleep", "12000"},
				},
			},
		},
	}
}

func CreateBusyboxTool(name string, labels, nodeSelector map[string]string, ctx *utils.TestContext) (*v1.Pod, error) {
	busyboxPod := generateBusybox(name, labels, nodeSelector)
	podURL := ctx.Cfg.K8SMasterForKubeEdge + constants.AppHandler
	podBytes, err := json.Marshal(busyboxPod)
	if err != nil {
		utils.Fatalf("Marshalling body failed: %v", err)
		return nil, err
	}
	err = handlePostRequest2K8s(podURL, podBytes)
	if err != nil {
		utils.Fatalf("Frame HTTP request to k8s failed: %v", err)
		return nil, err
	}
	gomega.Expect(err).To(gomega.BeNil())
	// wait pod ready
	//busyboxPodURL := podURL + "/" + name
	// TODO use labelselector to complte GetPod， not nodeName
	var pod *v1.Pod
	var podList v1.PodList
	for i := 0; i < 5; i++ {
		utils.Infof("Get Busybox tool pod round: %v", i)
		time.Sleep(5 * time.Second)
		podList, err = GetPodByLabels(labels, ctx)
		if err != nil {
			utils.Infof("GetPodByLabels failed: %v", err)
			continue
		}
		if len(podList.Items) == 0 {
			continue
		}
		pod = &podList.Items[0]
		break
	}
	if pod == nil {
		return nil, fmt.Errorf("can not get busybox tools pod")
	}
	utils.WaitforPodsRunning(ctx.Cfg.KubeConfigPath, podList, 240*time.Second)
	return pod, nil
}

func CleanBusyBoxTool(name string, ctx *utils.TestContext) error {
	podURL := ctx.Cfg.K8SMasterForKubeEdge + constants.AppHandler
	resp, err := utils.SendHTTPRequest(http.MethodDelete, podURL+"/"+name)
	if err != nil {
		utils.Fatalf("HTTP request is failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not equal to %d", http.StatusOK)
	}
	return nil
}

// generate hostname service object
func generateApplication(config *ApplicationConfig) (*appsv1.Deployment, *v1.Service) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GenDeploymentNameFromUID(config.Name),
			Labels:    config.Labels,
			Namespace: defaultNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(config.Replica); return &i }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: config.Labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: config.Labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            config.Name,
							Image:           config.ImageURL,
							ImagePullPolicy: "IfNotPresent",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: config.ContainerPort,
									HostPort:      config.HostPort,
								},
							},
						},
					},
					NodeSelector: config.NodeSelector,
				},
			},
		},
	}
	service := generateService(config.Name, config.Labels, config.ServicePortName,
		config.ServicePort, config.ServiceProtocol, config.ServiceTargetPort)

	return deployment, service
}

func GenServiceNameFromUID(uid string) string {
	return uid + "-svc"
}

func GenDeploymentNameFromUID(uid string) string {
	return uid
}

func GenDestinationRuleNameFromUID(uid string) string {
	return GenServiceNameFromUID(uid)
}

type ApplicationConfig struct {
	Name              string
	ImageURL          string
	NodeSelector      map[string]string
	Labels            map[string]string
	Replica           int
	HostPort          int32
	ContainerPort     int32
	ServicePortName   string
	ServicePort       int32
	ServiceProtocol   v1.Protocol
	ServiceTargetPort intstr.IntOrString
	Ctx               *utils.TestContext
}


func createApplication(config *ApplicationConfig) error {
	deployment, service := generateApplication(config)

	deployURL := config.Ctx.Cfg.K8SMasterForKubeEdge + constants.DeploymentHandler
	deployBytes, err := json.Marshal(deployment)
	if err != nil {
		utils.Fatalf("Marshalling body failed: %v", err)
		return err
	}
	err = handlePostRequest2K8s(deployURL, deployBytes)
	if err != nil {
		utils.Fatalf("Frame HTTP request to k8s failed: %v", err)
		return err
	}

	time.Sleep(5 * time.Second)
	// wait deployment ready
	// TODO use api/v1/namespaces/kube-system/pods?labelSelector=app%3Dhostname-edge
	podlist, err := GetPodByLabels(config.Labels, config.Ctx)
	if err != nil {
		utils.Fatalf("GetPods failed: %v", err)
		return err
	}
	utils.WaitforPodsRunning(config.Ctx.Cfg.KubeConfigPath, podlist, 240*time.Second)

	serviceURL := config.Ctx.Cfg.K8SMasterForKubeEdge + ServiceHandler
	serviceBytes, err := json.Marshal(service)
	if err != nil {
		utils.Fatalf("Marshalling body failed: %v", err)
		return err
	}
	err = handlePostRequest2K8s(serviceURL, serviceBytes)
	if err != nil {
		utils.Fatalf("Frame HTTP request to k8s failed: %v", err)
		return err
	}
	return nil
}

func CleanupApplication(name string, ctx *utils.TestContext) error {
	deploymentURL := ctx.Cfg.K8SMasterForKubeEdge + constants.DeploymentHandler
	resp, err := utils.SendHTTPRequest(http.MethodDelete, deploymentURL+"/"+GenDeploymentNameFromUID(name))
	if err != nil {
		utils.Fatalf("HTTP request is failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not equal to %d", http.StatusOK)
	}

	serviceURL := ctx.Cfg.K8SMasterForKubeEdge + ServiceHandler
	resp, err = utils.SendHTTPRequest(http.MethodDelete, serviceURL+"/"+GenServiceNameFromUID(name))
	if err != nil {
		utils.Fatalf("HTTP request is failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not equal to %d", http.StatusOK)
	}
	return nil
}

func CreateHostnameApplication(config *ApplicationConfig) error {
	return createApplication(config)
}

func CreateTCPReplyEdgemeshApplication(config *ApplicationConfig) error {
	return createApplication(config)
}


type LoadBalanceTesterApplicationConfig struct {
	Name         string
	Replica      int
	PodNames     []string // len(PodNames) == Replica
	NodeSelector map[string]string
	Labels       map[string]string
	ServicePort  int32
	Ctx          *utils.TestContext

	servicePortName string
	serviceProtocol v1.Protocol
	serviceTargetPort intstr.IntOrString
}

func generateHostnamePod(name string, hostPort int32, labels, nodeSelector map[string]string) *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: labels,
			Namespace: defaultNamespace,
		},
		Spec: v1.PodSpec{
			NodeSelector: nodeSelector,
			Containers: []v1.Container{
				{
					Name: "hostname",
					Image: "kevindavis/serve_hostname:latest",
					ImagePullPolicy: "IfNotPresent",
					Ports: []v1.ContainerPort{
						{
							ContainerPort: 9376,
							HostPort: hostPort,
						},
					},
				},
			},
		},
	}
}


// hostname pod will print podname
func CreateLoadBalanceTesterApplication(config *LoadBalanceTesterApplicationConfig) error {
	// generate pods and service
	var pods []*v1.Pod
	var service *v1.Service
	hostPort := int32(9376)
	for i := 0; i < config.Replica; i++ {
		pods = append(pods, generateHostnamePod(config.PodNames[i], hostPort, config.Labels, config.NodeSelector))
		hostPort += 1
	}
	config.servicePortName = "http-0"
	config.serviceProtocol = "TCP"
	config.serviceTargetPort = intstr.IntOrString{IntVal: 9376}
	service = generateService(config.Name, config.Labels, config.servicePortName, config.ServicePort, config.serviceProtocol,
		config.serviceTargetPort)

	// do http request
	podURL := config.Ctx.Cfg.K8SMasterForKubeEdge + constants.AppHandler
	for i := 0; i < len(pods); i++ {
		podBytes, err := json.Marshal(pods[i])
		if err != nil {
			utils.Fatalf("Marshalling body failed: %v", err)
			return err
		}
		err = handlePostRequest2K8s(podURL, podBytes)
		if err != nil {
			utils.Fatalf("Frame HTTP request to k8s failed: %v", err)
			return err
		}
	}
	time.Sleep(5*time.Second)
	podlist, err := GetPodByLabels(config.Labels, config.Ctx)
	if err != nil {
		utils.Fatalf("GetPodByLabels failed, %v", err)
		return err
	}
	utils.WaitforPodsRunning(config.Ctx.Cfg.KubeConfigPath, podlist, 240*time.Second)
	serviceBytes, err := json.Marshal(service)
	serviceURL := config.Ctx.Cfg.K8SMasterForKubeEdge + ServiceHandler
	if err != nil {
		utils.Fatalf("Marshalling body failed: %v", err)
		return err
	}
	err = handlePostRequest2K8s(serviceURL, serviceBytes)
	if err != nil {
		utils.Fatalf("Frame HTTP request to k8s failed: %v", err)
		return err
	}
	return nil
}


func generateService(name string, selector map[string]string, portName string, port int32,
	protocol v1.Protocol, targetPort intstr.IntOrString) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: GenServiceNameFromUID(name),
			Namespace: defaultNamespace,
		},
		Spec: v1.ServiceSpec{
			Selector: selector,
			Ports: []v1.ServicePort{
				{
					Name: portName,
					Port: port,
					Protocol: protocol,
					TargetPort: targetPort,
				},
			},
		},
	}
}