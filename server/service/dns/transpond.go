package dns

import (
	"context"
	"fmt"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dnsserver"
)

func (d *DnsService) setTransferCfgGather(cfg []*pb.TransferInfo, operType pb.OperType) ([]*OperateGather, error) {
	commonPrefix := d.KeyPrefix + "domain_manager/common/transpond_config/"
	prefix := d.KeyPrefix + "domain_manager/transpond_config/"

	var strIndex, tmpPrefix string
	oper := make([]*OperateGather, 0, 30)
	var tmpOper *OperateGather
	tmpOperType := int32(operType)

	for i := 0; i < len(cfg); i++ {
		strIndex = strconv.FormatUint(cfg[i].Id, 10)
		tmpPrefix = prefix + "config_index/" + strIndex

		if operType == pb.OperType_MOD {
			tmpOper = new(OperateGather)
			tmpOper.OperType = -1
			tmpOper.Key = tmpPrefix
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
		} else if operType == pb.OperType_DEL {
			tmpOper = new(OperateGather)
			tmpOper.OperType = tmpOperType
			tmpOper.Key = tmpPrefix
			tmpOper.Value = ""
			oper = append(oper, tmpOper)

			tmpOper = new(OperateGather)
			tmpOper.OperType = tmpOperType
			tmpOper.Key = commonPrefix + "config_index/" + strIndex
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
			continue
		}

		//domain name
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/domain_name"
		tmpOper.Value = cfg[i].Domain
		oper = append(oper, tmpOper)

		//IP
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/IP"
		tmpOper.Value = cfg[i].Ip
		oper = append(oper, tmpOper)

		//notes
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/notes"
		tmpOper.Value = cfg[i].Reference
		oper = append(oper, tmpOper)

		//status
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/status"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].Enable)
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (d *DnsService) OperateTransfer(ctx context.Context, req *pb.TransferInfos) (*pb.RespStatus, error) {
	logger.Debug("Enter OperateTransfer.")
	result := &pb.RespStatus{}

	if req == nil {
		result.Code = ErrorList[2].ErrCode
		result.Msg = ErrorList[2].ErrDesc
		return result, nil
	}

	oper, _ := d.setTransferCfgGather(req.Ev, req.OperateType)

	if d.OperateEtcdKv(oper) != nil {
		result.Code = ErrorList[0].ErrCode
		result.Msg = ErrorList[0].ErrDesc
	}

	logger.Debug("Exit.")

	return result, nil
}

func (d *DnsService) queryTransferInfo(kv clientv3.KV, key, index string) (*pb.TransferInfo, error) {
	info := new(pb.TransferInfo)
	info.Id, _ = strconv.ParseUint(index, 10, 64)
	key += index

	//domain_name
	tmpKey := key + "/domain_name"
	gresp, err := kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Domain = string(gresp.Kvs[0].Value)
	}

	//IP
	tmpKey = key + "/IP"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Ip = string(gresp.Kvs[0].Value)
	}

	//notes
	tmpKey = key + "/notes"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Reference = string(gresp.Kvs[0].Value)
	}

	//status
	tmpKey = key + "/status"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Enable, _ = strconv.ParseBool(string(gresp.Kvs[0].Value))
	}

	return info, nil
}

func (d *DnsService) GetTransfer(ctx context.Context, req *pb.ReqStatus) (*pb.TransferInfos, error) {
	logger.Debug("Enter GetTransfer.")

	if req == nil {
		logger.Debug("Input Parameter ReqStatus is nil.")
		return nil, nil
	}

	kvc := clientv3.NewKV(d.Db)
	var index []uint64
	if len(req.Id) > 0 && req.Id[0] == 0 {
		prefix := d.KeyPrefix + "domain_manager/common/transpond_config/config_index"
		index = d.queryAllId(kvc, prefix)
	} else {
		index = append(index, req.Id...)
	}

	prefix := d.KeyPrefix + "domain_manager/transpond_config/config_index/"

	var transferInfo []*pb.TransferInfo
	for i := 0; i < len(index); i++ {
		strIndex := strconv.FormatUint(index[i], 10)
		info, err := d.queryTransferInfo(kvc, prefix, strIndex)
		if err != nil {
			return nil, err
		}
		transferInfo = append(transferInfo, info)
	}

	resp := &pb.TransferInfos{
		Ev: transferInfo,
	}

	return resp, nil
}

func (d *DnsService) EnableTransfer(ctx context.Context, req *pb.EnableInfos) (*pb.RespStatus, error) {
	logger.Debug("Enter EnableTransfer.")

	result := &pb.RespStatus{}

	if req == nil {
		result.Code = ErrorList[2].ErrCode
		result.Msg = ErrorList[2].ErrDesc
		return result, nil
	}

	strIndex := strconv.FormatUint(req.Id, 10)

	strKey := d.KeyPrefix + "domain_manager/transpond_config/config_index/" +
		strIndex + "/status"
	reply, err := d.SetSingleKeyToDB(strKey, fmt.Sprintf("%v", req.Enable))
	if err != nil {
		result.Code = reply.Code
		result.Msg = reply.Msg
	}

	logger.Debug("Exit.")
	return result, nil
}
