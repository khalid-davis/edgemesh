package tunnel

import (
	"github.com/kubeedge/beehive/pkg/core"
	"github.com/kubeedge/edgemesh/pkg/apis/componentconfig/edgemesh/v1alpha1"
	"github.com/kubeedge/edgemesh/pkg/common/certificate"
	"github.com/kubeedge/edgemesh/pkg/common/modules"
	"github.com/kubeedge/edgemesh/pkg/tunnel/config"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/edgemesh/pkg/common/constants"
	gatewayConfig "github.com/kubeedge/edgemesh/pkg/networking/edgegateway/config"
	discoveryConfig "github.com/kubeedge/edgemesh/pkg/networking/servicediscovery/config"
	"github.com/kubeedge/edgemesh/pkg/tunnel/agentaddr"
	"github.com/kubeedge/edgemesh/pkg/tunnel/tunnelagent"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type Tunnel struct {
	certManager certificate.CertManager
	enable      bool
}

func NewTunnel(enable bool) *Tunnel {
	return &Tunnel{
		enable:      enable,
	}
}

func Register(tl *v1alpha1.Tunnel) {
	config.InitConfigure(tl)
	core.Register(NewTunnel(tl.Enable))
}

func (t *Tunnel) Name() string {
	return modules.AgentTunnelModuleName
}

func (t *Tunnel) Group() string {
	return modules.AgentTunnelGroupName
}

func (t *Tunnel) Enable() bool {
	return t.enable
}

func (t *Tunnel) Start() {
	certificateConfig := certificate.TunnelCertificate{
		Heartbeat:          config.Config.Heartbeat,
		TLSCAFile:          config.Config.TLSCAFile,
		TLSCertFile:        config.Config.TLSCertFile,
		TLSPrivateKeyFile:  config.Config.TLSPrivateKeyFile,
		Token:              config.Config.Token,
		HTTPServer:         config.Config.HTTPServer,
		RotateCertificates: config.Config.RotateCertificates,
		HostnameOverride:   config.Config.HostnameOverride,
	}
	t.certManager = certificate.NewCertManager(certificateConfig, config.Config.NodeName)
	t.certManager.Start()

	go tunnelagent.StartTunnelAgent()

	for {
		select {
		case <-beehiveContext.Done():
			klog.Warning("EdgeMesh stop")
			return
		default:
		}
		msg, err := beehiveContext.Receive(modules.AgentTunnelModuleName)
		if err != nil {
			klog.Warningf("Module %s receive msg error %v", modules.AgentTunnelModuleName, err)
			continue
		}
		klog.Infof("Module %s get message: %T", modules.AgentTunnelModuleName, msg)
		process(msg)
	}

	// TODO ifRotationDone() ????, 后面要添加这个东西，如果证书轮换了，要重新进行连接
}

func process(msg model.Message) {
	resource := msg.GetResource()
	switch resource {
	case constants.ResourceTypeSecret:
		if discoveryConfig.Config.Enable || gatewayConfig.Config.Enable {
			handleSecretMessage(msg)
		}
	}
}

func handleSecretMessage(msg model.Message) {
	secret, ok := msg.GetContent().(*v1.Secret)
	if !ok {
		klog.Warningf("object type: %T unsupported", secret)
		return
	}
	if secret.GetNamespace() == agentaddr.NewPeerAgentAddr().SecretNameSpace() && secret.GetName() == agentaddr.NewPeerAgentAddr().SecretName() {
		operation := msg.GetOperation()
		switch operation {
		case model.InsertOperation, model.UpdateOperation:
			agentaddr.NewPeerAgentAddr().Reset(secret.Data)
		}
	}
}