package dhcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

func (ds *DhcpService) setOptionsGather(cfg []*pb.OptionsInfo, operType pb.OperationType) ([]*OperateGather, error) {
	commonPrefix := ds.KeyPrefix + "address_manager/common/options"
	prefix := ds.KeyPrefix + "address_manager/options"
	var strIndex, tmpPrefix string
	oper := make([]*OperateGather, 0, 15)
	var tmpOper *OperateGather
	tmpOperType := int32(operType)

	for i := 0; i < len(cfg); i++ {
		strIndex = strconv.FormatUint(cfg[i].Index, 10)
		tmpPrefix = prefix + "/config_index/" + strIndex

		if operType == pb.OperationType_modify_type {
			tmpOper = new(OperateGather)
			tmpOper.OperType = -1
			tmpOper.Key = tmpPrefix
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
		} else if operType == pb.OperationType_delete_type {
			tmpOper = new(OperateGather)
			tmpOper.OperType = tmpOperType
			tmpOper.Key = tmpPrefix
			tmpOper.Value = ""
			oper = append(oper, tmpOper)

			tmpOper = new(OperateGather)
			tmpOper.OperType = tmpOperType
			tmpOper.Key = commonPrefix + "/config_index/" + strIndex
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
			continue
		}

		//option id
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/option_id"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].OptionId)
		oper = append(oper, tmpOper)

		//option name
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/option_name"
		tmpOper.Value = cfg[i].OptionName
		oper = append(oper, tmpOper)

		//option value
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/option_value"
		tmpOper.Value = cfg[i].OptionValue
		oper = append(oper, tmpOper)

		//protocol type
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/protocol_type"
		tmpOper.Value = fmt.Sprintf("%d", cfg[i].ProtocolType)
		oper = append(oper, tmpOper)

		//subnet id
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/subnet_id"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].SubnetId)
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + "/config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (ds *DhcpService) SetOptions(ctx context.Context, in *pb.OptionsReq) (*pb.RespResult, error) {
	logger.Debug("Enter SetOptions.")
	result := &pb.RespResult{}

	oper, _ := ds.setOptionsGather(in.Info, in.OperType)
	if ds.OperateEtcdKv(oper) != nil {
		result.ResultCode = ErrorList[0].ErrCode
		result.Description = ErrorList[0].ErrDesc
	}
	logger.Debug("Exit.")

	return result, nil
}

func (ds *DhcpService) queryOptionCfg(kv clientv3.KV, key, index string) (*pb.OptionsInfo, error) {
	gresp, err := kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	info := new(pb.OptionsInfo)
	info.Index, _ = strconv.ParseUint(index, 10, 64)

	for _, ev := range gresp.Kvs {
		tmpKey := key + "/option_id"
		if string(ev.Key) == tmpKey {
			tmpOptionId, _ := strconv.Atoi(string(ev.Value))
			info.OptionId = uint32(tmpOptionId)
			continue
		}

		tmpKey = key + "/option_name"
		if string(ev.Key) == tmpKey {
			info.OptionName = string(ev.Value)
			continue
		}

		tmpKey = key + "/option_value"
		if string(ev.Key) == tmpKey {
			info.OptionValue = string(ev.Value)
			continue
		}

		tmpKey = key + "/protocol_type"
		if string(ev.Key) == tmpKey {
			tmpProtoType, _ := strconv.Atoi(string(ev.Value))
			info.ProtocolType = pb.ProtocolType(tmpProtoType)
			continue
		}

		tmpKey = key + "/subnet_id"
		if string(ev.Key) == tmpKey {
			info.SubnetId, _ = strconv.ParseUint(string(ev.Value), 10, 64)
			continue
		}
	}

	return info, nil
}

func (ds *DhcpService) GetOptions(ctx context.Context, in *pb.ReqStatus) (*pb.OptionsRsp, error) {
	logger.Debug("Enter GetOptions.")
	rsp := &pb.OptionsRsp{}
	if in == nil {
		return nil, errors.New("Input Parameter is error.")
	}

	var cfg []*pb.OptionsInfo
	kvc := clientv3.NewKV(ds.Db)
	prefix := ds.KeyPrefix + "address_manager/options/"
	if in.Index != 0 {
		strIndex := strconv.FormatUint(in.Index, 10)
		prefix += "config_index/" + strIndex
		tmpCfg, err := ds.queryOptionCfg(kvc, prefix, strIndex)
		if err == nil {
			cfg = append(cfg, tmpCfg)
		}

	} else if in.Index == 0 {
		commPrefix := ds.KeyPrefix + "address_manager/common/fingerprint/config_index/"
		var strKey, strIndex string
		gresp, _ := kvc.Get(context.TODO(), commPrefix, clientv3.WithPrefix())
		for _, ev := range gresp.Kvs {
			strKey = string(ev.Key)
			strIndex = strKey[len(commPrefix):]
			tmpKey := prefix
			tmpKey += "config_index/" + strIndex
			tmpCfg, err := ds.queryOptionCfg(kvc, tmpKey, strIndex)
			if err == nil {
				cfg = append(cfg, tmpCfg)
			}
		}
	}

	rsp.Info = append(rsp.Info, cfg...)
	logger.Debug("Exit.")
	return rsp, nil
}
