package dns

import (
	"context"
	"fmt"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dnsserver"
)

func (d *DnsService) setRecursionCfgGather(cfg *pb.RecursionInfo) ([]*OperateGather, error) {
	prefix := d.KeyPrefix + "domain_manager/recursion_resolver"

	oper := make([]*OperateGather, 0, 30)
	var tmpOper *OperateGather
	tmpOperType := int32(pb.OperType_ADD)

	//status
	tmpOper = new(OperateGather)
	tmpOper.OperType = tmpOperType
	tmpOper.Key = prefix + "/status"
	tmpOper.Value = fmt.Sprintf("%v", cfg.RecursionEnable)
	oper = append(oper, tmpOper)

	//dns46_enable
	tmpOper = new(OperateGather)
	tmpOper.OperType = tmpOperType
	tmpOper.Key = prefix + "/dns46_enable"
	tmpOper.Value = fmt.Sprintf("%v", cfg.Dns46Enable)
	oper = append(oper, tmpOper)

	//dns64_enable
	tmpOper = new(OperateGather)
	tmpOper.OperType = tmpOperType
	tmpOper.Key = prefix + "/dns64_enable"
	tmpOper.Value = fmt.Sprintf("%v", cfg.Dns64Enable)
	oper = append(oper, tmpOper)

	//dns64_synthall
	tmpOper = new(OperateGather)
	tmpOper.OperType = tmpOperType
	tmpOper.Key = prefix + "/dns64_synthall"
	tmpOper.Value = fmt.Sprintf("%v", cfg.Dns64Synthall)
	oper = append(oper, tmpOper)

	//dns64_prefix
	tmpOper = new(OperateGather)
	tmpOper.OperType = tmpOperType
	tmpOper.Key = prefix + "/dns64_prefix"
	tmpOper.Value = cfg.Dns64Prefix
	oper = append(oper, tmpOper)

	//JumpAddr
	tmpOper = new(OperateGather)
	tmpOper.OperType = tmpOperType
	tmpOper.Key = prefix + "/jump_address"
	tmpOper.Value = cfg.JumpAddr
	oper = append(oper, tmpOper)

	return oper, nil
}

func (d *DnsService) UpdateRecursion(ctx context.Context, req *pb.RecursionInfo) (*pb.RespStatus, error) {
	logger.Debug("Enter UpdateRecursion.")
	result := &pb.RespStatus{}

	if req == nil {
		result.Code = ErrorList[2].ErrCode
		result.Msg = ErrorList[2].ErrDesc
		return result, nil
	}

	oper, _ := d.setRecursionCfgGather(req)

	if d.OperateEtcdKv(oper) != nil {
		result.Code = ErrorList[0].ErrCode
		result.Msg = ErrorList[0].ErrDesc
	}

	logger.Debug("Exit.")
	return result, nil
}

func (d *DnsService) GetRecursion(ctx context.Context, req *pb.ReqStatus) (*pb.RecursionInfo, error) {
	logger.Debug("Enter GetRecursion.")

	kv := clientv3.NewKV(d.Db)

	prefix := d.KeyPrefix + "domain_manager/recursion_resolver"

	info := new(pb.RecursionInfo)
	// status
	tmpKey := prefix + "/status"
	gresp, err := kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.RecursionEnable, _ = strconv.ParseBool(string(gresp.Kvs[0].Value))
	}

	//dns46_enable
	tmpKey = prefix + "/dns46_enable"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Dns46Enable, _ = strconv.ParseBool(string(gresp.Kvs[0].Value))
	}

	//dns64_enable
	tmpKey = prefix + "/dns64_enable"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Dns64Enable, _ = strconv.ParseBool(string(gresp.Kvs[0].Value))
	}

	//dns64_synthall
	tmpKey = prefix + "/dns64_synthall"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Dns64Synthall, _ = strconv.ParseBool(string(gresp.Kvs[0].Value))
	}

	//dns64_prefix
	tmpKey = prefix + "/dns64_prefix"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Dns64Prefix = string(gresp.Kvs[0].Value)
	}

	//jump_address
	tmpKey = prefix + "/jump_address"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.JumpAddr = string(gresp.Kvs[0].Value)
	}

	return info, nil
}
