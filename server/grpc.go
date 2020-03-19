package server

import (
	"log"
	"net"

	"google.golang.org/grpc"
	pb "reyzar.com/server-api/server/service/dhcp"
	"reyzar.com/server-api/server/service/dns"
)

func (s *Server) initRPC() {
	var err error
	s.grpcListener, err = net.Listen("tcp", s.config.ServerAddr)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(s.config.DhcpAgentAddr); i++ {
		conn, err := grpc.Dial(s.config.DhcpAgentAddr[i], grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		s.DhcpAgentConn = append(s.DhcpAgentConn, conn)
	}

	s.grpcServer = grpc.NewServer()
	pb.RegisterDhcpService(s.grpcServer, s.etcd, s.DhcpAgentConn, s.config.KeyPrefix)
	dns.RegisterDnsService(s.grpcServer, s.etcd, s.config.KeyPrefix)
}

func (s *Server) startRPC() {
	err := s.grpcServer.Serve(s.grpcListener)
	if err != nil {
		log.Fatal(err)
	}
}
