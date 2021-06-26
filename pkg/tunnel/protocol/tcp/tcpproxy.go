package tcp

import (
	"context"
	"github.com/kubeedge/edgemesh/pkg/tunnel/agentaddr"
	tcp_pb "github.com/kubeedge/edgemesh/pkg/tunnel/protocol/tcp/pb"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-msgio/protoio"
	"io"
	"k8s.io/klog/v2"
	"net"
	"time"
)

var TCPProxyProtocol protocol.ID = "/libp2p/tcpproxy/1.0.0"

type TCPProxyService struct {
	host host.Host
}

func NewTCPProxyService(h host.Host) *TCPProxyService{
	return &TCPProxyService{
		host: h,
	}
}

// used by server
func (tp *TCPProxyService) ProxyStreamHandler(s network.Stream) {
	// todo use peerID to get nodeName
	klog.Infof("Get a new stream from %s", s.Conn().RemotePeer().String())
	streamWriter := protoio.NewDelimitedWriter(s)
	streamReader := protoio.NewDelimitedReader(s, 4 * 1024)

	msg := new(tcp_pb.TCPProxy)
	if err := streamReader.ReadMsg(msg); err != nil {
		klog.Errorf("Read msg from %s err: err", s.Conn().RemotePeer().String(), err)
		s.Reset()
		return
	}
	if msg.GetType() != tcp_pb.TCPProxy_CONNECT {
		klog.Errorf("Read msg from %s type should be CONNECT", s.Conn().RemotePeer().String())
		s.Reset()
		return
	}

	targetIP := msg.GetIp()
	targetPort := msg.GetPort()
	targetAddr := &net.TCPAddr{
		IP:   net.ParseIP(targetIP),
		Port: int(targetPort),
	}
	klog.Infof("l4 proxy get tcp server address: %v", targetAddr)

	var proxyClient net.Conn
	var err error
	// todo retry time use const
	for i := 0; i < 5; i++ {
		proxyClient, err = net.DialTimeout("tcp", targetAddr.String(), 5 * time.Second)
		if err == nil {
			break
		}
	}

	msg.Reset()
	if err != nil {
		klog.Errorf("l4 proxy connect to %s:%d err: %v", targetIP, targetPort, err)
		msg.Type = tcp_pb.TCPProxy_FAILED.Enum()
		if err := streamWriter.WriteMsg(msg); err != nil {
			klog.Errorf("Write msg to %s err: %v", s.Conn().RemotePeer().String(), err)
			s.Reset()
			return
		}
		return
	}

	msg.Type = tcp_pb.TCPProxy_SUCCESS.Enum()
	if err := streamWriter.WriteMsg(msg); err != nil {
		klog.Errorf("Write msg to %s err: %v", s.Conn().RemotePeer().String(), err)
		s.Reset()
		return
	}
	msg.Reset()

	go io.Copy(proxyClient, s)
	go io.Copy(s, proxyClient)
}

func (tp *TCPProxyService) GetProxyStream(targetNodeName, targetIP string, targetPort int) (io.ReadWriteCloser, error) {
	klog.Infof("Get %s proxy stream between %s", TCPProxyProtocol, targetNodeName)
	destAddr, err := agentaddr.NewPeerAgentAddr().Get(targetNodeName)
	if err != nil {
		klog.Errorf("Get %s addr err: %v", targetNodeName, err)
		return nil, err
	}
	destInfo, err := peer.AddrInfoFromP2pAddr(destAddr)
	if err != nil {
		klog.Errorf("Transfer multiAddr %s to peer info err: %v", destAddr.String(), err)
		return nil, err
	}

	connNum := tp.host.Network().ConnsToPeer(destInfo.ID)
	if len(connNum) >= 2 {
		klog.Infof("Data transfer between %s is p2p mode", targetNodeName)
	} else {
		klog.Infof("Try to hole punch with %s", targetNodeName)
		// todo add retry
		err = tp.host.Connect(context.Background(), *destInfo)
		if err != nil {
			klog.Errorf("connect to %s err: %v", targetNodeName, err)
			return nil, err
		}
		klog.Infof("Data transfer between %s is realy mode", targetNodeName)
	}

	stream, err := tp.host.NewStream(context.Background(), destInfo.ID, TCPProxyProtocol)
	if err != nil {
		klog.Errorf("New Stream between %s err: %v", targetNodeName, err)
		// todo if the reason is no conn, add retry connect to target node
		return nil, err
	}
	klog.Infof("Get %s proxy stream between %s successful", TCPProxyProtocol, targetNodeName)

	streamWriter := protoio.NewDelimitedWriter(stream)
	// todo msg size use const
	streamReader := protoio.NewDelimitedReader(stream, 4 * 1024)

	port := int32(targetPort)
	msg := &tcp_pb.TCPProxy{
		Type:                 tcp_pb.TCPProxy_CONNECT.Enum(),
		Nodename:             &targetNodeName,
		Ip:                   &targetIP,
		Port:                 &port,
	}

	if err = streamWriter.WriteMsg(msg); err != nil {
		klog.Errorf("Write conn msg to %s err: %v", targetNodeName, err)
		stream.Reset()
		return nil, err
	}
	msg.Reset()

	if err = streamReader.ReadMsg(msg); err != nil {
		klog.Errorf("Read conn result msg to %s err: %v", targetNodeName, err)
		stream.Reset()
		return nil, err
	}
	if msg.GetType() != tcp_pb.TCPProxy_SUCCESS {
		// todo targetnode write detailed fail info back
		klog.Errorf("%s dial %s:%d err: %v", targetNodeName, targetIP, targetPort, err)
		stream.Reset()
		return nil, err
	}
	msg.Reset()

	return stream, nil
}