package dhcp_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"reyzar.com/server-api/pkg/rpc/dhcpclient"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

const address1 = ":50056"
const address2 = ":50055"

const address = "127.0.0.1:50057"

type testServer struct {
}

func (ts *testServer) GetDeviceInfo(ctx context.Context, in *dhcpclient.GetDeviceInfoReq) (*dhcpclient.GetDeviceInfoRsp, error) {
	log.Println("Enter GetDeviceInfo.")
	log.Println("Input Para: ", in)
	rsp := &dhcpclient.GetDeviceInfoRsp{
		IpAddr:          "10.2.1.17",
		CpuUsageRate:    0.55,
		MemoryUsageRate: 0.32,
		DiskUsageRate:   0.11,
	}

	return rsp, nil
}

func newGrpcServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	tsServer := &testServer{}
	dhcpclient.RegisterDevicePerformanceServer(server, tsServer)

	server.Serve(listener)
}

func TestGetDeviceInfo(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewDevicePerformanceClient(conn)

	req := &pb.GetDeviceInfoReq{
		IpAddr: []string{"10.2.1.17", "10.2.1.18"},
	}

	reply, err := c.GetDeviceInfo(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(reply)
}

func TestStartGrpc(t *testing.T) {
	log.Println("##")
	go newGrpcServer(address1)
	go newGrpcServer(address2)
	ticker := time.NewTicker(time.Duration(20) * time.Second)
	for {
		select {
		case <-ticker.C:
		}
	}
}
