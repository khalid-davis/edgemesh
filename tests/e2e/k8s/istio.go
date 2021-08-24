package k8s

import (
	"encoding/json"
	"github.com/kubeedge/kubeedge/tests/e2e/utils"
	apiv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	destinationRuleHandler = "/apis/networking.istio.io/v1alpha3/namespaces/default/destinationrules"
)

var loadBalanceMap = map[string]apiv1alpha3.LoadBalancerSettings_SimpleLB {
	"RoundRobin": apiv1alpha3.LoadBalancerSettings_ROUND_ROBIN,
}

func CreateDestinationRule(uid string, loadBalancer string, ctx *utils.TestContext) error {
	destinationRule := &v1alpha3.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GenDestinationRuleNameFromUID(uid),
			Namespace: defaultNamespace,
		},
		Spec: apiv1alpha3.DestinationRule{
			Host: GenDestinationRuleNameFromUID(uid),
			TrafficPolicy: &apiv1alpha3.TrafficPolicy{
				LoadBalancer: &apiv1alpha3.LoadBalancerSettings{
					LbPolicy: &apiv1alpha3.LoadBalancerSettings_Simple {
						Simple: loadBalanceMap[loadBalancer],
					},
				},
			},
		},
	}

	destinationRuleURL := ctx.Cfg.K8SMasterForKubeEdge + destinationRuleHandler
	destinationRuleBytes, err := json.Marshal(destinationRule)
	if err != nil {
		utils.Fatalf("Marshalling body failed: %v", err)
		return err
	}
	err = handlePostRequest2K8s(destinationRuleURL, destinationRuleBytes)
	if err != nil {
		utils.Fatalf("Frame HTTP request to k8s failed: %v", err)
		return err
	}
	return nil

}
