package traffic

import (
	"fmt"
	"github.com/kubeedge/edgemesh/tests/e2e/k8s"
	"github.com/kubeedge/kubeedge/tests/e2e/constants"
	"github.com/kubeedge/kubeedge/tests/e2e/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"time"
)

var DeploymentTestTimerGroup = utils.NewTestTimerGroup()

var _ = Describe("Traffic", func() {
	var UID string
	var testTimer *utils.TestTimer
	var testDescription GinkgoTestDescription

	Context("Test same lan features", func() {
		BeforeEach(func() {
			// Get current test description
			testDescription = CurrentGinkgoTestDescription()
			// Start test timer
			testTimer = DeploymentTestTimerGroup.NewTestTimer(testDescription.TestText)
			time.Sleep(10 * time.Second)
		})
		AfterEach(func() {
			// End test timer
			testTimer.End()
			// Print result
			testTimer.PrintResult()
			var podlist corev1.PodList
			var deploymentList appsv1.DeploymentList
			err := utils.GetDeployments(&deploymentList, ctx.Cfg.K8SMasterForKubeEdge+constants.DeploymentHandler)
			Expect(err).To(BeNil())
			for _, deployment := range deploymentList.Items {
				if deployment.Name == UID {
					labels := deployment.Spec.Selector.MatchLabels
					podlist, err = k8s.GetPodByLabels(labels, ctx)
					Expect(err).To(BeNil())
					err := k8s.CleanupApplication(UID, ctx)
					Expect(err).To(BeNil())
				}
			}
			utils.Infof("podlist len %v", len(podlist.Items))
			utils.CheckPodDeleteState(ctx.Cfg.K8SMasterForKubeEdge+constants.AppHandler, podlist)
			utils.PrintTestcaseNameandStatus()
		})

		// whether the service discovery feature work
		It("E2E_TRAFFIC_SAME_LAN_SERVICE_DISCOVERY: Create hostname service and busybox pod to check the service "+
			"discovery feature", func() {
			// 1. start hostname application
			UID = "hostname-applicatioin-" + utils.GetRandomString(5)
			nodeSelector := map[string]string{"lan": "edge-lan-01"}
			servicePort := int32(12345)
			err := CreateHostnameApplication(UID, nodeSelector, servicePort, 1, ctx)
			Expect(err).To(BeNil())

			// 2. get clusterIP
			service, err := k8s.GetService(k8s.GenServiceNameFromUID(UID), ctx)
			Expect(err).To(BeNil())
			clusterIP := service.Spec.ClusterIP

			// 3. use busybox pod exec dig <hostname-svc>.<namespace>
			// use sys.call docker exec -it <busybox-id> dig hostname-svc.default A +noall +answer +tries=10
			// s := "docker exec 980f466def1d dig hostname-svc.default"
			domain := k8s.GenServiceNameFromUID(UID) + "." + defaultNamespace
			command := fmt.Sprintf("docker exec -i %s dig %s A +noall +answer +tries=5", busyboxToolContainerID, domain)
			utils.Infof("command %v", command)
			// TODO always the first time can not fetch the result or maybe just because the cluster is new
			var outStr, resultIP string
			for i := 0; i < 20; i++ {
				time.Sleep(5 * time.Second)
				utils.Infof("dig command retry round: %v", i)
				outStr, err = k8s.CallSysCommand(command)
				if err != nil {
					continue
				}
				//4. fetch resultIP and check
				resultIP = k8s.FetchIPFromDigOutput(outStr, domain)
				if resultIP != "" {
					break
				}
			}
			Expect(err).To(BeNil())
			utils.Infof("OutStr: %v, domain: %v", outStr, domain)
			utils.Infof("Expect IP %v, domain IP: %v", resultIP, clusterIP)
			Expect(resultIP).To(Equal(clusterIP))
		})


		// whether the http traffic governance feature work
		It("E2E_TRAFFIC_SAME_LAN_HTTP_TRAFFIC_Governance", func() {
			// 1. start hostname application
			UID = "hostname-application-" + utils.GetRandomString(5)
			nodeSelector := map[string]string{"lan": "edge-lan-01"}
			servicePort := int32(12345)
			err := CreateHostnameApplication(UID, nodeSelector, servicePort, 1, ctx)
			Expect(err).To(BeNil())

			// 2. use busybox pod exec curl <hostname-svc>.<namespace>:<svc-port>
			domain := k8s.GenServiceNameFromUID(UID) + "." + defaultNamespace
			command := fmt.Sprintf("docker exec -i %s curl -w %%{http_code} %s:%d", busyboxToolContainerID, domain, servicePort)
			utils.Infof("command %v", command)
			// wait for the hostname application be ready, the pod running no equals to the server ready
			var outStr string
			var statusCode int
			for i := 0; i <= 20; i++ {
				time.Sleep(5 * time.Second)
				utils.Infof("curl command retry round: %v", i)
				outStr, err = k8s.CallSysCommand(command)
				if err != nil {
					continue
				}
				_, statusCode = k8s.FetchHostnameAndStatusCodeFromOutput(outStr)
				if statusCode != http.StatusOK {
					continue
				}
				break
			}
			Expect(err).To(BeNil())
			utils.Infof("outStr: %v", outStr)

			// 3. check the http status equal OK
			Expect(statusCode).To(Equal(http.StatusOK))
		})


		//whether the tcp traffic governance feature work
		It("E2E_TRAFFIC_SAME_LAN_TCP_TRAFFIC_Governance", func() {
			// 1. start tcp-reply-edgemesh application
			UID = "tcp-reply-edgemesh-application-" + utils.GetRandomString(5)
			nodeSelector := map[string]string{"lan": "edge-lan-01"}
			servicePort := int32(12345)
			err := CreateTCPReplyEdgemeshApplication(UID, nodeSelector, servicePort, 1, ctx)
			Expect(err).To(BeNil())

			// 2. use busybox pod exec curl <tcp-reply-edgemesh-svc>.<namespace> <svc-port>
			domain := k8s.GenServiceNameFromUID(UID) + "." + defaultNamespace
			command := fmt.Sprintf("docker exec -i %s curl %s:%d", busyboxToolContainerID, domain, servicePort)
			utils.Infof("command %v", command)
			//wait for the hostname application be ready, the pod running no equals to the server ready
			var outStr string
			for i := 0; i <= 20; i++ {
				time.Sleep(5 * time.Second)
				utils.Infof("curl command retry round: %v", i)
				outStr, err = k8s.CallSysCommand(command)
				if err == nil {
					break
				}
			}
			Expect(err).To(BeNil())
			utils.Infof("outStr: %v", outStr)

			// 3. check whether the reply equals to "edgemesh"
			reply := k8s.FetchTCPReplyFromOutput(outStr)
			utils.Infof("reply: %v", reply)
			Expect(reply).To(Equal("edgemesh"))
		})


		// 启动两个实例，配置轮询策略：访问10次，他们的分布应该是：[1010101010]，[0101010101]
		// can not test until https://github.com/kubeedge/edgemesh/pull/16 merged
		//// whether the load balance feature work using destinationRule
		//It("E2E_TRAFFIC_SAME_LAN_LOADBALANCE", func() {
		//	// 1. create two hostname pod and service
		//	UID = "lb-hostname-application-" + utils.GetRandomString(5)
		//	nodeSelector := map[string]string{"lan": "edge-lan-01"}
		//	podNames := []string{"lb-hostname-01", "lb-hostname-02"}
		//	servicePort := int32(12345)
		//	err := CreateLoadBalanceTesterApplication(UID, podNames, nodeSelector, servicePort, ctx)
		//	Expect(err).To(BeNil())
		//
		//	// 2. create destinationRule with RoundRobin strategy
		//	err = k8s.CreateDestinationRule(UID, "RoundRobin", ctx)
		//	Expect(err).To(BeNil())
		//
		//	// 3. curl the service 10 times, check the result is match for RoundRobin strategy
		//	domain := k8s.GenServiceNameFromUID(UID) + "." + defaultNamespace
		//	time.Sleep(10 * time.Second)
		//	count := 10
		//	rsl := make([]string, count)
		//	for i := 0; i < count; i++ {
		//		command := fmt.Sprintf("docker exec -i %s curl -w %%{http_code} %s:%d", busyboxToolContainerID, domain, servicePort)
		//		utils.Infof("command %v", command)
		//		outStr, err := k8s.CallSysCommand(command)
		//		Expect(err).To(BeNil())
		//		utils.Infof("outStr: %v", outStr)
		//
		//		hostName, statusCode := k8s.FetchHostnameAndStatusCodeFromOutput(outStr)
		//		Expect(statusCode).To(Equal(http.StatusOK))
		//		rsl[i] = hostName
		//	}
		//	var first, second string
		//	if podNames[0] == rsl[0] {
		//		first, second = podNames[0], podNames[1]
		//	} else {
		//		first, second = podNames[1], podNames[0]
		//	}
		//	expectRsl := make([]string, count)
		//	for i := 0; i < count/2; i++ {
		//		rsl[i] = first
		//		rsl[i+1] = second
		//	}
		//
		//	utils.Infof("expect: %v, get: %v", expectRsl, rsl)
		//	Expect(expectRsl).To(Equal(rsl))
		//})
		//
		//// whether the consistent load balance strategy work
		//It("E2E_TRAFFIC_SAME_LAN_LOADBALANCE_CONSISTENT", func() {
		//
		//})
	})
})
