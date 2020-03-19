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

func (ds *DhcpService) setFingerprintGather(cfg []*pb.FingerprintInfo, operType pb.OperationType) ([]*OperateGather, error) {
	commonPrefix := ds.KeyPrefix + "address_manager/common/fingerprint"
	prefix := ds.KeyPrefix + "address_manager/fingerprint"
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

		// ttl
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/ttl"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].Ttl)
		oper = append(oper, tmpOper)

		//options
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/options"
		tmpOper.Value = cfg[i].Options
		oper = append(oper, tmpOper)

		//option id
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/option_id"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].OptionId)
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

		//message type
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/message_type"
		tmpOper.Value = fmt.Sprintf("%d", cfg[i].MessageType)
		oper = append(oper, tmpOper)

		//supplier
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/supplier"
		tmpOper.Value = cfg[i].Supplier
		oper = append(oper, tmpOper)

		//system_name
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/system_name"
		tmpOper.Value = cfg[i].SystemName
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + "/config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (ds *DhcpService) SetFingerprint(ctx context.Context, in *pb.FingerprintReq) (*pb.RespResult, error) {
	logger.Debug("Enter SetFingerprint.")
	result := &pb.RespResult{}

	oper, _ := ds.setFingerprintGather(in.CfgInfo, in.OperType)

	if ds.OperateEtcdKv(oper) != nil {
		result.ResultCode = ErrorList[0].ErrCode
		result.Description = ErrorList[0].ErrDesc
	}

	logger.Debug("exit.")
	return result, nil
}

func (ds *DhcpService) queryFingerprintCfg(kv clientv3.KV, key, index string) (*pb.FingerprintInfo, error) {
	gresp, err := kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	info := new(pb.FingerprintInfo)
	info.Index, _ = strconv.ParseUint(index, 10, 64)

	for _, ev := range gresp.Kvs {
		tmpKey := key + "/ttl"
		if string(ev.Key) == tmpKey {
			tmpTtl, _ := strconv.Atoi(string(ev.Value))
			info.Ttl = uint32(tmpTtl)
			continue
		}

		tmpKey = key + "/options"
		if string(ev.Key) == tmpKey {
			info.Options = string(ev.Value)
			continue
		}

		tmpKey = key + "/option_id"
		if string(ev.Key) == tmpKey {
			tmpOptionId, _ := strconv.Atoi(string(ev.Value))
			info.OptionId = uint32(tmpOptionId)
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

		tmpKey = key + "/message_type"
		if string(ev.Key) == tmpKey {
			tmpMessageType, _ := strconv.Atoi(string(ev.Value))
			info.MessageType = int32(tmpMessageType)
			continue
		}

		tmpKey = key + "/supplier"
		if string(ev.Key) == tmpKey {
			info.Supplier = string(ev.Value)
			continue
		}

		tmpKey = key + "/system_name"
		if string(ev.Key) == tmpKey {
			info.SystemName = string(ev.Value)
			continue
		}
	}

	return info, nil
}

func (ds *DhcpService) GetFingerprint(ctx context.Context, in *pb.ReqStatus) (*pb.FingerprintRsp, error) {
	logger.Debug("Enter GetFingerprint.")
	rsp := &pb.FingerprintRsp{}
	if in == nil {
		return nil, errors.New("Input Parameter is error.")
	}

	var cfg []*pb.FingerprintInfo
	kvc := clientv3.NewKV(ds.Db)
	prefix := ds.KeyPrefix + "address_manager/fingerprint/"
	if in.Index != 0 {
		strIndex := strconv.FormatUint(in.Index, 10)
		prefix += "config_index/" + strIndex
		tmpCfg, err := ds.queryFingerprintCfg(kvc, prefix, strIndex)
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
			tmpCfg, err := ds.queryFingerprintCfg(kvc, tmpKey, strIndex)
			if err == nil {
				cfg = append(cfg, tmpCfg)
			}
		}
	}

	rsp.CfgInfo = append(rsp.CfgInfo, cfg...)
	logger.Debug("Exit.")
	return rsp, nil
}
