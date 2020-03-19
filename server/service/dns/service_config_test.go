package dns_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"google.golang.org/grpc"
	pb "reyzar.com/server-api/pkg/rpc/dnsserver"
)

const address = "127.0.0.1:50057"

func TestUpdateSysConf(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDnsManagerClient(conn)

	conf := new(pb.SysConf)
	conf.Port = 53
	conf.CnamePriority = true
	conf.TcpEnable = true

	var info []*pb.ServiceIPInfo
	tmpInfo := new(pb.ServiceIPInfo)
	tmpInfo.Ip = "10.2.21.1"
	info = append(info, tmpInfo)

	tmpInfo = new(pb.ServiceIPInfo)
	tmpInfo.Ip = "10.2.21.2"
	info = append(info, tmpInfo)

	req := &pb.SysConfs{
		Sys: conf,
		Ip:  info,
	}

	rsp, err := c.UpdateSysConf(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}

func TestGetSysConf(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDnsManagerClient(conn)

	req := &pb.ReqStatus{}
	rsp, err := c.GetSysConf(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}
