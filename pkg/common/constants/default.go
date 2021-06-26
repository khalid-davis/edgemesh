package constants

import v1 "k8s.io/api/core/v1"

// Resources
const (
	// Certificates
	DefaultConfigDir = "/etc/kubeedge/config/"

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
	ResourceTypeSecret     = "secret"
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
	ServerDefaultCAFile = "/etc/kubeedge/edgemesh/server/ca/rootCA.crt"
	ServerDefaultCertFile = "/etc/kubeedge/edgemesh/server/certs/server.crt"
	ServerDefaultKeyFile = "/etc/kubeedge/edgemesh/server/certs/server.key"
	AgentDefaultCAFile = "/etc/kubeedge/edgemesh/agent/ca/rootCA.crt"
	AgentDefaultCertFile = "/etc/kubeedge/edgemesh/agent/certs/server.crt"
	AgentDefaultKeyFile = "/etc/kubeedge/edgemesh/agent/certs/server.key"
)