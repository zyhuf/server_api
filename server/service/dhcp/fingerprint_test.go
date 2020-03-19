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

func ConstructFingerprint() []*pb.FingerprintInfo {
	var info []*pb.FingerprintInfo

	tmpInfo := new(pb.FingerprintInfo)
	tmpInfo.Index = 6001
	// tmpInfo.Ttl = 100
	// tmpInfo.Options = "12,12,12"
	// tmpInfo.OptionId = 33
	// tmpInfo.OptionValue = "192.168.1.3"
	// tmpInfo.Supplier = "睿哲"
	// tmpInfo.SystemName = "安卓"
	info = append(info, tmpInfo)

	tmpInfo = new(pb.FingerprintInfo)
	tmpInfo.Index = 6002
	// tmpInfo.Ttl = 200
	// tmpInfo.Options = "15,16,17"
	// tmpInfo.OptionId = 35
	// tmpInfo.OptionValue = "120"
	// tmpInfo.Supplier = "华为"
	// tmpInfo.SystemName = "曚昽"
	info = append(info, tmpInfo)

	return info
}

func TestSetFingerprint(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewFingerprintManagerClient(conn)

	//构造新增
	cfg := ConstructFingerprint()
	req := &pb.FingerprintReq{
		OperType: pb.OperationType_delete_type,
		CfgInfo:  cfg,
	}

	reply, err := c.SetFingerprint(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(reply)
}

func TestGetFingerprint(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewFingerprintManagerClient(conn)

	req := &pb.ReqStatus{
		Index: 6001,
	}

	reply, err := c.GetFingerprint(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(reply)
}
