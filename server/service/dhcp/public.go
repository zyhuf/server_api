package dhcp

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

var ErrorList = [...]struct {
	ErrCode int32
	ErrDesc string
}{
	{1, "Access etcd failed."},
	{2, "The index is invalid."},
	{3, "The list type is invalid."},
	{4, "Invalid parameter."},
	{5, "The record does not exist."},
	{6, "The parameter TYPE is error."},
}

type DhcpService struct {
	Db             *clientv3.Client
	RdhcpAgentConn []*grpc.ClientConn
	KeyPrefix      string
}

type OperateGather struct {
	OperType int32
	Key      string
	Value    string
}

const GRPC_RESULT_RATE = 5000

func RegisterDhcpService(server *grpc.Server, etcd *clientv3.Client, conn []*grpc.ClientConn, keyPrefix string) {
	service := &DhcpService{
		Db:             etcd,
		RdhcpAgentConn: conn,
		KeyPrefix:      keyPrefix,
	}

	pb.RegisterServiceConfigServer(server, service)
	pb.RegisterAccessControlServer(server, service)
	pb.RegisterFailoverManagerServer(server, service)
	pb.RegisterAuthManagerServer(server, service)
	pb.RegisterAddressReconcileManagerServer(server, service)
	pb.RegisterDevicePerformanceServer(server, service)
	pb.RegisterFingerprintManagerServer(server, service)
	pb.RegisterOptionsManagerServer(server, service)
}

func (ds *DhcpService) OperateEtcdKv(oper []*OperateGather) error {
	var err error

	var op []clientv3.Op
	var opDel []clientv3.Op
	for j := 0; j < len(oper); j++ {
		if oper[j].OperType == int32(pb.OperationType_add_type) ||
			oper[j].OperType == int32(pb.OperationType_modify_type) {
			//log.Println("key:", oper[j].Key, " value:", oper[j].Value)
			tmpOp := clientv3.OpPut(oper[j].Key, oper[j].Value)
			op = append(op, tmpOp)
		} else if oper[j].OperType == int32(pb.OperationType_delete_type) {
			//log.Println("key:", oper[j].Key, " value:", oper[j].Value)
			tmpOp := clientv3.OpDelete(oper[j].Key, clientv3.WithPrefix())
			op = append(op, tmpOp)
		} else if oper[j].OperType == -1 {
			tmpOp := clientv3.OpDelete(oper[j].Key, clientv3.WithPrefix())
			opDel = append(opDel, tmpOp)
		}
	}

	kvc := clientv3.NewKV(ds.Db)
	if len(opDel) != 0 {
		_, err = kvc.Txn(context.TODO()).Then(opDel...).Commit()
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	_, err = kvc.Txn(context.TODO()).Then(op...).Commit()
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (ds *DhcpService) SetSingleKeyToDB(key, value string) (*pb.RespResult, error) {
	result := &pb.RespResult{}

	kvc := clientv3.NewKV(ds.Db)
	gresp, err := kvc.Get(context.TODO(), key)
	if err != nil {
		logger.Error(err)
		result.ResultCode = ErrorList[0].ErrCode
		result.Description = ErrorList[0].ErrDesc
		return result, nil
	}

	if len(gresp.Kvs) == 0 {
		result.ResultCode = ErrorList[4].ErrCode
		result.Description = ErrorList[4].ErrDesc
		return result, nil
	}

	_, err = kvc.Put(context.TODO(), key, value)
	if err != nil {
		logger.Error(err)
		result.ResultCode = ErrorList[0].ErrCode
		result.Description = ErrorList[0].ErrDesc
		return result, nil
	}

	return result, nil
}
