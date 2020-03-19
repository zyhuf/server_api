package dhcp_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"google.golang.org/grpc"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

const address = "127.0.0.1:50057"

func constructOption() []*pb.OptionsInfo {
	var rsp []*pb.OptionsInfo
	info := new(pb.OptionsInfo)
	info.Index = 1101
	// info.OptionId = 33
	// info.OptionName = "static route"
	// info.OptionValue = "172.16.128.2"
	// info.ProtocolType = pb.ProtocolType_DHCPv4_ProtocolType
	// info.SubnetId = 1001
	rsp = append(rsp, info)

	info = new(pb.OptionsInfo)
	info.Index = 1102
	// info.OptionId = 35
	// info.OptionName = "ARP cache timeout"
	// info.OptionValue = "300"
	// info.ProtocolType = pb.ProtocolType_DHCPv6_ProtocolType
	// info.SubnetId = 1002
	rsp = append(rsp, info)

	return rsp
}

func TestSetOptions(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewOptionsManagerClient(conn)

	cfg := constructOption()
	req := &pb.OptionsReq{
		OperType: pb.OperationType_delete_type,
		Info:     cfg,
	}

	reply, err := c.SetOptions(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(reply)
}

func TestGetOptions(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewOptionsManagerClient(conn)
	req := &pb.ReqStatus{
		Index: 1101,
	}

	reply, err := c.GetOptions(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(reply)
}
