package constants

import v1 "k8s.io/api/core/v1"

// Resources
const (
	// Certificates
	//DefaultConfigDir = "/etc/kubeedge/config/"
	// TODO delete
	DefaultConfigDir = "D:\\workspace\\gocode\\gomodule\\local-conf"

	// Config
	DefaultKubeContentType         = "application/vnd.kubernetes.protobuf"
	DefaultKubeConfig              = "/root/.kube/config"
	DefaultKubeNamespace           = v1.NamespaceAll
	DefaultKubeQPS                 = 100.0
	DefaultKubeBurst               = 200
	DefaultKubeUpdateNodeFrequency = 20

	// Controller
	DefaultServiceEventBuffer         = 1
	DefaultEndpointsEventBuffer       = 1
	DefaultDestinationRuleEventBuffer = 1
	DefaultGatewayEventBuffer         = 1

	// Resource
	ResourceTypeService     = "service"
	ResourceTypeEndpoints   = "endpoints"
	ResourceDestinationRule = "destinationRule"
	ResourceTypeGateway     = "gateway"
)

const (
	NodeName = "NodeName"
	CloudCoreToken = "CloudCoreToken"
)

const (
	// Tunnel modules
	DefaultCAURL        = "/ca.crt"
	DefaultAgentCertURL = "/agent.crt"
	DefaultHostnameOverride = "default-agent-node"
	//DefaultCAFile = "/etc/kubeedge/edgemesh/ca/rootCA.crt"
	//DefaultCertFile = "/etc/kubeedge/edgemesh/certs/server.crt"
	//DefaultKeyFile = "/etc/kubeedge/edgemesh/certs/server.key"
	// TODO delete
	DefaultCAFile = "D:\\workspace\\gocode\\gomodule\\local-conf\\edgemesh-agent\\ca\\rootCA.crt"
	DefaultCertFile = "D:\\workspace\\gocode\\gomodule\\local-conf\\edgemesh-agent\\certs\\server.crt"
	DefaultKeyFile = "D:\\workspace\\gocode\\gomodule\\local-conf\\edgemesh-agent\\certs\\server.key"

)