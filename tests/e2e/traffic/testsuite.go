package traffic

import (
	"github.com/kubeedge/edgemesh/tests/e2e/k8s"
	"github.com/kubeedge/kubeedge/tests/e2e/utils"
	"k8s.io/apimachinery/pkg/util/intstr"
	"math/rand"
)

var (
	defaultNamespace = "default"
	ServiceHandler   = "/api/v1/namespaces/default/services"
)

func CreateHostnameApplication(uid string, nodeSelector map[string]string, servicePort int32, replica int, ctx *utils.TestContext) error {
	imgURL := "k8s.gcr.io/serve_hostname:latest"
	labels := map[string]string{"app": "hostname-edge"}
	config := &k8s.ApplicationConfig{
		Name:            uid,
		ImageURL:        imgURL,
		NodeSelector:    nodeSelector,
		Labels:          labels,
		Replica:         replica,
		HostPort:        9376 + rand.Int31n(20),
		ContainerPort:   9376,
		// TODO wait for the zb pr
		// ServicePortName: "http-0",
		ServicePortName: "tcp-0",
		ServicePort:     servicePort,
		ServiceProtocol: "TCP",
		ServiceTargetPort: intstr.IntOrString{IntVal: 9376},
		Ctx:             ctx,
	}
	return k8s.CreateHostnameApplication(config)
}


func CreateTCPReplyEdgemeshApplication(uid string, nodeSelector map[string]string, servicePort int32, replica int, ctx *utils.TestContext) error {
	imgURL := "kevindavis/tcp-reply-edgemesh:v1.0"
	labels := map[string]string{"app": "tcp-reply-edgemesh-edge"}
	config := &k8s.ApplicationConfig{
		Name:              uid,
		ImageURL:          imgURL,
		NodeSelector:      nodeSelector,
		Labels:            labels,
		Replica:           replica,
		HostPort:          9001,
		ContainerPort:     9001,
		ServicePortName:   "tcp-0",
		ServicePort:       servicePort,
		ServiceProtocol:   "TCP",
		ServiceTargetPort: intstr.IntOrString{IntVal: 9001},
		Ctx:               ctx,
	}
	return k8s.CreateTCPReplyEdgemeshApplication(config)
}

func CreateLoadBalanceTesterApplication(uid string, podNames []string, nodeSelector map[string]string,
	servicePort int32, ctx *utils.TestContext) error {
	labels := map[string]string{"app": "hostname-lb-edge"}
	config := &k8s.LoadBalanceTesterApplicationConfig{
		Name:         uid,
		Replica:      2,
		PodNames:     podNames,
		NodeSelector: nodeSelector,
		Labels:       labels,
		ServicePort:  servicePort,
		Ctx:          ctx,
	}
	return k8s.CreateLoadBalanceTesterApplication(config)
}