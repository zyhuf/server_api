package dhcp

import (
	"context"
	"fmt"
	"log"
	"testing"

	"google.golang.org/grpc"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

const address = "127.0.0.1:50057"

func TestSetAddressInspectCfg(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewAddressReconcileManagerClient(conn)

	req := &pb.InspectCfgReq{
		InspectCycle: 300,
	}

	reply, err := c.SetAddressInspectCfg(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(reply)
}
