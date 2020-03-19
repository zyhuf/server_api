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

func ConstructForwardInfos() *pb.ForwardInfos {
	var info []*pb.ForwardInfo

	tmpInfo := new(pb.ForwardInfo)
	tmpInfo.Id = 1001
	tmpInfo.Domain = "baidu1.com"
	tmpInfo.Reference = "域名百度1"
	info = append(info, tmpInfo)

	tmpInfo = new(pb.ForwardInfo)
	tmpInfo.Id = 1002
	tmpInfo.Domain = "reyzar.com"
	tmpInfo.Reference = "域名睿哲"
	info = append(info, tmpInfo)

	reply := &pb.ForwardInfos{
		OperateType: pb.OperType_ADD,
		Ev:          info,
	}

	return reply
}

func TestOperateForward(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDnsManagerClient(conn)

	req := ConstructForwardInfos()
	rsp, err := c.OperateForward(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}

func TestGetForward(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDnsManagerClient(conn)

	var id []uint64
	id = append(id, 1001)
	id = append(id, 1002)
	req := &pb.ReqStatus{
		Id: id,
	}
	rsp, err := c.GetForward(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}

func ConstructDomainInfos() *pb.DomainInfos {
	var info []*pb.DomainInfo

	tmpInfo := new(pb.DomainInfo)
	tmpInfo.Id = 2001
	// tmpInfo.ForId = 1001
	// tmpInfo.Name = "www"
	// tmpInfo.Type = "A"
	// tmpInfo.IspId = 1
	// tmpInfo.Value = "39.156.69.79"
	// tmpInfo.Ttl = 300
	// tmpInfo.Mx = 2
	// tmpInfo.Enable = true
	info = append(info, tmpInfo)

	tmpInfo = new(pb.DomainInfo)
	tmpInfo.Id = 2002
	// tmpInfo.ForId = 1002
	// tmpInfo.Name = "www"
	// tmpInfo.Type = "A"
	// tmpInfo.IspId = 1
	// tmpInfo.Value = "218.13.22.46"
	// tmpInfo.Ttl = 300
	// tmpInfo.Mx = 2
	// tmpInfo.Enable = true
	info = append(info, tmpInfo)

	reply := &pb.DomainInfos{
		OperateType: pb.OperType_DEL,
		Ev:          info,
	}

	return reply
}

func TestOperateDomain(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDnsManagerClient(conn)

	req := ConstructDomainInfos()
	rsp, err := c.OperateDomain(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}

func TestGetDomain(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDnsManagerClient(conn)

	req := &pb.ForwardRefer{
		DomainId:  2001,
		ForwardId: 1001,
	}

	rsp, err := c.GetDomain(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}

func TestEnableDomain(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDnsManagerClient(conn)

	refer := new(pb.ForwardRefer)
	refer.DomainId = 2001
	refer.ForwardId = 1001

	req := &pb.ForwardEnableMsg{
		Refer:  refer,
		Enable: false,
	}

	rsp, err := c.EnableDomain(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}

func TestGetForwardStatus(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDnsManagerClient(conn)

	var info []*pb.ForStatusInfo

	tmpInfo := new(pb.ForStatusInfo)
	refer := new(pb.ForwardRefer)
	refer.DomainId = 1001
	refer.ForwardId = 2001
	tmpInfo.Refer = refer
	info = append(info, tmpInfo)

	tmpInfo = new(pb.ForStatusInfo)
	refer = new(pb.ForwardRefer)
	refer.DomainId = 1002
	refer.ForwardId = 2002
	tmpInfo.Refer = refer
	info = append(info, tmpInfo)

	req := &pb.ForwardStatusInfos{
		Ev: info,
	}

	rsp, err := c.GetForwardStatus(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(rsp)
}
