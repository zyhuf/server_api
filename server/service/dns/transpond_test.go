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

func ConstructTransferInfos() *pb.TransferInfos {
	var info []*pb.TransferInfo

	tmpInfo := new(pb.TransferInfo)
	tmpInfo.Id = 3001
	tmpInfo.Domain = "example1.com"
	tmpInfo.Ip = "192.168.2.1"
	tmpInfo.Enable = true
	tmpInfo.Reference = "测试1"
	info = append(info, tmpInfo)

	tmpInfo = new(pb.TransferInfo)
	tmpInfo.Id = 3002
	tmpInfo.Domain = "example2.com"
	tmpInfo.Ip = "192.168.2.3"
	tmpInfo.Enable = false
	tmpInfo.Reference = "测试2"
	info = append(info, tmpInfo)

	result := &pb.TransferInfos{
		OperateType: pb.OperType_ADD,
		Ev:          info,
	}

	return result
}

func TestOperateTransfer(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDnsManagerClient(conn)
	req := ConstructTransferInfos()

	rsp, err := c.OperateTransfer(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}

func TestGetTransfer(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	c := pb.NewDnsManagerClient(conn)

	var id []uint64
	id = append(id, 3001)
	id = append(id, 3002)

	req := &pb.ReqStatus{
		Id: id,
	}

	rsp, err := c.GetTransfer(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}

func TestEnableTransfer(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	c := pb.NewDnsManagerClient(conn)

	req := &pb.EnableInfos{
		Id:     3001,
		Enable: true,
	}

	rsp, err := c.EnableTransfer(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}
