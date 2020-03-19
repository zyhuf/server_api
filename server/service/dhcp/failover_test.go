package dhcp_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"testing"

	"google.golang.org/grpc"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

const address = "127.0.0.1:50057"

func AddOrModSeverCfg() []*pb.SetServerCfgReq {
	var req []*pb.SetServerCfgReq

	tmpReq := new(pb.SetServerCfgReq)
	tmpReq.OperType = pb.OperationType_add_type

	var info []*pb.ServerCfgInfo
	tmpInfo := new(pb.ServerCfgInfo)
	tmpInfo.Index = 3001
	tmpInfo.ServerType = pb.ServerType_master_server
	tmpInfo.LocalAddress = "241c::1"
	tmpInfo.LocalPort = 5001
	tmpInfo.PeerAddress = "242c::2"
	tmpInfo.PeerPort = 5002
	tmpInfo.MonitorTime = 100
	tmpInfo.MaxUpdateTimes = 10
	tmpInfo.MaxLoadBalancingTime = 200
	tmpInfo.AutoUpdateLeaseTime = 20
	tmpInfo.SeparateDigit = 128
	info = append(info, tmpInfo)

	tmpInfo = new(pb.ServerCfgInfo)
	tmpInfo.Index = 3002
	tmpInfo.ServerType = pb.ServerType_slave_server
	tmpInfo.LocalAddress = "241b::1"
	tmpInfo.LocalPort = 6001
	tmpInfo.PeerAddress = "242b::2"
	tmpInfo.PeerPort = 6002
	tmpInfo.MonitorTime = 300
	tmpInfo.MaxUpdateTimes = 20
	tmpInfo.MaxLoadBalancingTime = 300
	tmpInfo.AutoUpdateLeaseTime = 30
	tmpInfo.SeparateDigit = 64
	info = append(info, tmpInfo)

	tmpReq.ServerConfig = append(tmpReq.ServerConfig, info...)
	req = append(req, tmpReq)

	return req
}

func DelServerCfg() []*pb.SetServerCfgReq {
	var req []*pb.SetServerCfgReq

	tmpReq := new(pb.SetServerCfgReq)
	tmpReq.OperType = pb.OperationType_delete_type

	var info []*pb.ServerCfgInfo
	tmpInfo := new(pb.ServerCfgInfo)
	tmpInfo.Index = 3001
	tmpInfo.ServerType = pb.ServerType_master_server
	info = append(info, tmpInfo)

	tmpInfo = new(pb.ServerCfgInfo)
	tmpInfo.Index = 3002
	tmpInfo.ServerType = pb.ServerType_slave_server
	info = append(info, tmpInfo)

	tmpReq.ServerConfig = append(tmpReq.ServerConfig, info...)
	req = append(req, tmpReq)

	return req
}

func TestSetServerCfg(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewFailoverManagerClient(conn)

	stream, err := c.SetServerCfg(context.TODO())
	if err != nil {
		log.Fatalln(err)
	}

	//构造新增
	req := AddOrModSeverCfg()
	//req := DelServerCfg()
	log.Println("size:", len(req))
	for i := 0; i < len(req); i++ {
		tmpReq := &pb.SetServerCfgReq{
			OperType:     req[i].OperType,
			ServerConfig: req[i].ServerConfig,
		}

		err = stream.Send(tmpReq)
		if err != nil {
			log.Fatalln("##", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalln("###", err)
	}

	log.Println(reply)
}

func TestGetServerCfg(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewFailoverManagerClient(conn)

	req := &pb.GetServerCfgReq{
		Index: 3001,
		//ServerType: pb.ServerType_master_server,
	}

	stream, err := c.GetServerCfg(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln(err)
		}

		for i := 0; i < len(reply.RespServer); i++ {
			log.Println(reply.RespServer[i])
		}
	}
}
