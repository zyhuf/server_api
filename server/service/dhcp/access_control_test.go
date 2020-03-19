package dhcp_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"

	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

const address = "127.0.0.1:50057"

func AddOrModBlackAndWhiteList() []*pb.BlackAndWhiteListReq {
	var req []*pb.BlackAndWhiteListReq

	tmpReq := new(pb.BlackAndWhiteListReq)
	tmpReq.OperType = pb.OperationType_modify_type

	var info []*pb.BlackAndWhiteListInfo
	tmpInfo := new(pb.BlackAndWhiteListInfo)
	tmpInfo.Index = 1001
	tmpInfo.ListType = pb.ListType_black_list
	tmpInfo.MacAddress = "241c::1"
	tmpInfo.UserName = "平板电脑"
	tmpInfo.SubnetId = 1111
	tmpInfo.Notes = "张三可以使用"
	tmpInfo.Status = pb.Status_enable_status
	info = append(info, tmpInfo)

	tmpInfo = new(pb.BlackAndWhiteListInfo)
	tmpInfo.Index = 1002
	tmpInfo.ListType = pb.ListType_black_list
	tmpInfo.MacAddress = "241c::2"
	tmpInfo.UserName = "PC机"
	tmpInfo.SubnetId = 2222
	tmpInfo.Notes = "李四的"
	tmpInfo.Status = pb.Status_enable_status
	info = append(info, tmpInfo)

	tmpReq.ListInfo = append(tmpReq.ListInfo, info...)
	req = append(req, tmpReq)

	tmpReq = new(pb.BlackAndWhiteListReq)
	tmpReq.OperType = pb.OperationType_modify_type
	var info2 []*pb.BlackAndWhiteListInfo
	tmpInfo = new(pb.BlackAndWhiteListInfo)
	tmpInfo.Index = 2001
	tmpInfo.ListType = pb.ListType_white_list
	tmpInfo.MacAddress = "2002::1"
	tmpInfo.UserName = "华为手机"
	tmpInfo.SubnetId = 1234
	tmpInfo.Notes = "张三可以使用"
	tmpInfo.Status = pb.Status_enable_status
	info2 = append(info2, tmpInfo)

	tmpInfo = new(pb.BlackAndWhiteListInfo)
	tmpInfo.Index = 2002
	tmpInfo.ListType = pb.ListType_white_list
	tmpInfo.MacAddress = "2002::2"
	tmpInfo.UserName = "苹果电脑"
	tmpInfo.SubnetId = 1235
	tmpInfo.Notes = "李四的"
	tmpInfo.Status = pb.Status_enable_status
	info2 = append(info2, tmpInfo)

	tmpReq.ListInfo = append(tmpReq.ListInfo, info2...)
	req = append(req, tmpReq)

	return req
}

func DelBlackAndWhiteList() []*pb.BlackAndWhiteListReq {
	var req []*pb.BlackAndWhiteListReq

	tmpReq := new(pb.BlackAndWhiteListReq)
	tmpReq.OperType = pb.OperationType_delete_type

	var info []*pb.BlackAndWhiteListInfo
	tmpInfo := new(pb.BlackAndWhiteListInfo)
	tmpInfo.Index = 1001
	tmpInfo.ListType = pb.ListType_black_list
	info = append(info, tmpInfo)

	tmpInfo = new(pb.BlackAndWhiteListInfo)
	tmpInfo.Index = 1002
	tmpInfo.ListType = pb.ListType_black_list
	info = append(info, tmpInfo)

	tmpReq.ListInfo = append(tmpReq.ListInfo, info...)
	req = append(req, tmpReq)

	tmpReq = new(pb.BlackAndWhiteListReq)
	tmpReq.OperType = pb.OperationType_delete_type
	var info2 []*pb.BlackAndWhiteListInfo
	tmpInfo = new(pb.BlackAndWhiteListInfo)
	tmpInfo.Index = 2001
	tmpInfo.ListType = pb.ListType_white_list
	info2 = append(info2, tmpInfo)

	tmpInfo = new(pb.BlackAndWhiteListInfo)
	tmpInfo.Index = 2002
	tmpInfo.ListType = pb.ListType_white_list
	info2 = append(info2, tmpInfo)

	tmpReq.ListInfo = append(tmpReq.ListInfo, info2...)
	req = append(req, tmpReq)

	return req
}

func TestSetBlackAndWhiteList(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewAccessControlClient(conn)

	stream, err := c.SetBlackAndWhiteList(context.TODO())
	if err != nil {
		log.Fatalln(err)
	}

	//构造新增
	req := AddOrModBlackAndWhiteList()
	//req := DelBlackAndWhiteList()
	log.Println("size:", len(req))
	for i := 0; i < len(req); i++ {
		tmpReq := &pb.BlackAndWhiteListReq{
			OperType: req[i].OperType,
			ListInfo: req[i].ListInfo,
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

func TestGetBlackAndWhiteList(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewAccessControlClient(conn)

	req := &pb.GetBWListReq{
		//Index:    1002,
		ListType: pb.ListType_black_list,
	}

	stream, err := c.GetBlackAndWhiteList(context.TODO(), req)
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

		for i := 0; i < len(reply.RespList); i++ {
			log.Println(reply.RespList[i])
		}
	}
}

func TestSetBWListStatus(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewAccessControlClient(conn)

	req := &pb.StatusBWListReq{
		Index:    1001,
		ListType: pb.ListType_black_list,
		Status:   pb.Status_disable_stauts,
	}

	reply, err := c.SetBWListStatus(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(reply)
}

func TestTxn(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	kvc := clientv3.NewKV(cli)
	txn := kvc.Txn(context.Background())
	var op []clientv3.Op

	op2 := clientv3.OpDelete("/test", clientv3.WithPrefix())
	op = append(op, op2)

	op1 := clientv3.OpPut("/test/12", "123")
	op = append(op, op1)

	txn = txn.Then(op...)
	txn.Commit()

	gresp, err := kvc.Get(context.TODO(), "/test")
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range gresp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}

	//kvc.Txn(context.Background()).Then(clientv3.OpDelete("/test", clientv3.WithPrefix())).Commit()

}
