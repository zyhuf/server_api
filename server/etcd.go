package server

import (
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
)

func (s *Server) initEtcd() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   s.config.DbAddr,
		DialTimeout: 3 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
	}

	s.etcd = client
}
