package dns

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dnsserver"
)

var ErrorList = [...]struct {
	ErrCode int32
	ErrDesc string
}{
	{1001, "Access etcd failed."},
	{1002, "The index is invalid."},
	{1003, "Invalid parameter."},
	{1004, "The record does not exist."},
}

type OperateGather struct {
	OperType int32
	Key      string
	Value    string
}

type DnsService struct {
	Db        *clientv3.Client
	KeyPrefix string
}

func RegisterDnsService(server *grpc.Server, etcd *clientv3.Client, keyPrefix string) {
	service := &DnsService{
		Db:        etcd,
		KeyPrefix: keyPrefix,
	}

	pb.RegisterDnsManagerServer(server, service)
}

func (d *DnsService) OperateEtcdKv(oper []*OperateGather) error {
	var err error

	var op []clientv3.Op
	var opDel []clientv3.Op
	for j := 0; j < len(oper); j++ {
		if oper[j].OperType == int32(pb.OperType_ADD) ||
			oper[j].OperType == int32(pb.OperType_MOD) {
			//log.Println("key:", oper[j].Key, " value:", oper[j].Value)
			tmpOp := clientv3.OpPut(oper[j].Key, oper[j].Value)
			op = append(op, tmpOp)
		} else if oper[j].OperType == int32(pb.OperType_DEL) {
			//log.Println("key:", oper[j].Key, " value:", oper[j].Value)
			tmpOp := clientv3.OpDelete(oper[j].Key, clientv3.WithPrefix())
			op = append(op, tmpOp)
		} else if oper[j].OperType == -1 {
			tmpOp := clientv3.OpDelete(oper[j].Key, clientv3.WithPrefix())
			opDel = append(opDel, tmpOp)
		}
	}

	kvc := clientv3.NewKV(d.Db)
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

func (d *DnsService) SetSingleKeyToDB(key, value string) (*pb.RespStatus, error) {
	result := &pb.RespStatus{}

	kvc := clientv3.NewKV(d.Db)
	gresp, err := kvc.Get(context.TODO(), key)
	if err != nil {
		logger.Error(err)
		result.Code = ErrorList[0].ErrCode
		result.Msg = ErrorList[0].ErrDesc
		return result, nil
	}

	if len(gresp.Kvs) == 0 {
		result.Code = ErrorList[3].ErrCode
		result.Msg = ErrorList[3].ErrDesc
		return result, nil
	}

	_, err = kvc.Put(context.TODO(), key, value)
	if err != nil {
		logger.Error(err)
		result.Code = ErrorList[0].ErrCode
		result.Msg = ErrorList[0].ErrDesc
		return result, nil
	}

	return result, nil
}
