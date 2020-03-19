package server

import (
	"net"

	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
)

type Server struct {
	grpcListener  net.Listener
	grpcServer    *grpc.Server
	etcd          *clientv3.Client
	config        *Config
	DhcpAgentConn []*grpc.ClientConn
}

func NewServer(config *Config) (s *Server) {
	s = &Server{
		config: config,
	}
	return s
}

func (s *Server) Start() {
	s.initEtcd()
	s.initRPC()

	s.startRPC()
}
