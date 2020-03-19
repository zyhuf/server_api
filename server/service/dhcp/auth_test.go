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

func AddOrModAuthcfg() []*pb.AuthCfgReq {
	var req []*pb.AuthCfgReq

	tmpReq := new(pb.AuthCfgReq)
	tmpReq.OperType = pb.OperationType_add_type

	var info []*pb.AuthCfgInfo
	tmpInfo := new(pb.AuthCfgInfo)
	tmpInfo.Index = 3001
	tmpInfo.Type = pb.AuthType_perpetual_auth
	tmpInfo.MacAddress = "241d::1"
	tmpInfo.UserName = "平板电脑"
	tmpInfo.ExpireTime = 0
	tmpInfo.Notes = "张三可以使用"
	tmpInfo.Status = pb.Status_enable_status
	info = append(info, tmpInfo)

	tmpInfo = new(pb.AuthCfgInfo)
	tmpInfo.Index = 3002
	tmpInfo.Type = pb.AuthType_temporary_auth
	tmpInfo.MacAddress = "242d::1"
	tmpInfo.UserName = "手机"
	tmpInfo.ExpireTime = 600
	tmpInfo.ExpireDate = "2020-02-01 19:30:11"
	tmpInfo.Notes = "李四可以使用"
	tmpInfo.Status = pb.Status_enable_status
	info = append(info, tmpInfo)

	tmpReq.AuthConfig = append(tmpReq.AuthConfig, info...)
	req = append(req, tmpReq)

	return req
}

func DelAuthCfg() []*pb.AuthCfgReq {
	var req []*pb.AuthCfgReq

	tmpReq := new(pb.AuthCfgReq)
	tmpReq.OperType = pb.OperationType_delete_type

	var info []*pb.AuthCfgInfo
	tmpInfo := new(pb.AuthCfgInfo)
	tmpInfo.Index = 3001
	tmpInfo.Type = pb.AuthType_perpetual_auth
	info = append(info, tmpInfo)

	tmpInfo = new(pb.AuthCfgInfo)
	tmpInfo.Index = 3002
	tmpInfo.Type = pb.AuthType_temporary_auth
	info = append(info, tmpInfo)

	tmpReq.AuthConfig = append(tmpReq.AuthConfig, info...)
	req = append(req, tmpReq)

	return req
}

func TestSetAuthCfg(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewAuthManagerClient(conn)

	stream, err := c.SetAuthCfg(context.TODO())
	if err != nil {
		log.Fatalln(err)
	}

	//构造新增
	req := AddOrModAuthcfg()
	//req := DelAuthCfg()
	log.Println("size:", len(req))
	for i := 0; i < len(req); i++ {
		tmpReq := &pb.AuthCfgReq{
			OperType:   req[i].OperType,
			AuthConfig: req[i].AuthConfig,
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

func TestGetAuthCfg(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewAuthManagerClient(conn)

	req := &pb.GetAuthCfgReq{
		Index: 3001,
		Type:  pb.AuthType_perpetual_auth,
	}

	stream, err := c.GetAuthCfg(context.TODO(), req)
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

		for i := 0; i < len(reply.RespAuth); i++ {
			log.Println(reply.RespAuth[i])
		}
	}
}

func TestSetAuthStatus(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewAuthManagerClient(conn)

	req := &pb.StatusAuthReq{
		Index:  3001,
		Type:   pb.AuthType_perpetual_auth,
		Status: pb.Status_disable_stauts,
	}

	reply, err := c.SetAuthStatus(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(reply)
}

func ConstructDisableList() []*pb.DisableUserListInfo {
	var info []*pb.DisableUserListInfo
	tmpInfo := new(pb.DisableUserListInfo)
	tmpInfo.Index = 5001
	//tmpInfo.MacAddress = "12-12-13-ab-cd-11"
	//tmpInfo.UserName = "张三"
	info = append(info, tmpInfo)

	tmpInfo = new(pb.DisableUserListInfo)
	tmpInfo.Index = 5002
	//tmpInfo.MacAddress = "11-17-13-ab-cd-22"
	//tmpInfo.UserName = "李四"
	info = append(info, tmpInfo)

	return info
}

func TestSetDisableUserList(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewAuthManagerClient(conn)
	stream, err := c.SetDisableUserList(context.TODO())
	if err != nil {
		log.Fatalln(err)
	}

	//构造新增
	cfg := ConstructDisableList()
	tmpReq := &pb.DisableUserListReq{
		OperType: pb.OperationType_delete_type,
		CfgInfo:  cfg,
	}

	err = stream.Send(tmpReq)
	if err != nil {
		log.Fatalln("##", err)
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalln("###", err)
	}

	log.Println(reply)
}

func TestGetDisableUserList(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewAuthManagerClient(conn)

	req := &pb.ReqStatus{
		//Index: 5001,
	}

	stream, err := c.GetDisableUserList(context.TODO(), req)
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

		for i := 0; i < len(reply.CfgInfo); i++ {
			log.Println(reply.CfgInfo[i])
		}
	}
}
