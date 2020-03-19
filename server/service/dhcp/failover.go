package dhcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

func (ds *DhcpService) setServerCfgGather(cfg []*pb.ServerCfgInfo, operType pb.OperationType) ([]*OperateGather, error) {
	commonPrefix := ds.KeyPrefix + "address_manager/common/failover/"
	prefix := ds.KeyPrefix + "address_manager/failover/"

	var serverType, strIndex, tmpPrefix string
	oper := make([]*OperateGather, 0, 15)
	var tmpOper *OperateGather
	tmpOperType := int32(operType)

	for i := 0; i < len(cfg); i++ {
		if cfg[i].ServerType == pb.ServerType_master_server {
			serverType = "master_server"
		} else if cfg[i].ServerType == pb.ServerType_slave_server {
			serverType = "slave_server"
		} else {
			return nil, errors.New("The server type is Error.")
		}

		strIndex = strconv.FormatUint(cfg[i].Index, 10)
		tmpPrefix = prefix + serverType + "/config_index/" + strIndex

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
			tmpOper.Key = commonPrefix + serverType + "/config_index/" + strIndex
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
			continue
		}

		//local address
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/local_address"
		tmpOper.Value = cfg[i].LocalAddress
		oper = append(oper, tmpOper)

		//local port
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/local_port"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].LocalPort)
		oper = append(oper, tmpOper)

		//peer address
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/peer_address"
		tmpOper.Value = cfg[i].PeerAddress
		oper = append(oper, tmpOper)

		//peer port
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/peer_port"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].PeerPort)
		oper = append(oper, tmpOper)

		//监测对端是否失效时间间隔（s）
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/monitor_cycle"
		tmpOper.Value = strconv.FormatUint(cfg[i].MonitorTime, 10)
		oper = append(oper, tmpOper)

		//最大无限制更新(BNDUPD)次数
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/max_update_times"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].MaxUpdateTimes)
		oper = append(oper, tmpOper)

		//最大负载平衡时间（s）
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/max_load_balancing_time"
		tmpOper.Value = strconv.FormatUint(cfg[i].MaxLoadBalancingTime, 10)
		oper = append(oper, tmpOper)

		//peer之间未联系时自动更新lease时间（mclt）(s)
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/auto_update_lease_time"
		tmpOper.Value = strconv.FormatUint(cfg[i].AutoUpdateLeaseTime, 10)
		oper = append(oper, tmpOper)

		//分隔位（0~256）
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/separate_digit"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].SeparateDigit)
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + serverType + "/config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (ds *DhcpService) SetServerCfg(stream pb.FailoverManager_SetServerCfgServer) error {
	logger.Debug("Enter SetServerCfg.")
	result := &pb.RespResult{}

	var oper []*OperateGather
	var tmpOper []*OperateGather
	var err error

	for {
		cfg, err := stream.Recv()
		if err == io.EOF {
			logger.Debug("All messages had received.")
			break
		}

		if err != nil {
			logger.Error(err)
			return err
		}

		tmpOper, err = ds.setServerCfgGather(cfg.ServerConfig, cfg.OperType)
		if err != nil {
			logger.Error(err)
			result.ResultCode = ErrorList[5].ErrCode
			result.Description = ErrorList[5].ErrDesc
			err = stream.SendAndClose(result)
			logger.Debug("Exec SendAndClose, err=", err)
			return err
		}
		oper = append(oper, tmpOper...)
	}

	if ds.OperateEtcdKv(oper) != nil {
		result.ResultCode = ErrorList[0].ErrCode
		result.Description = ErrorList[0].ErrDesc
	}

	err = stream.SendAndClose(result)
	logger.Debug("Exec SendAndClose, err=", err)

	logger.Debug("Exit.")
	return err
}

func (ds *DhcpService) queryServerInfo(kv clientv3.KV, serverType pb.ServerType, key, index string) (*pb.ServerCfgInfo, error) {
	gresp, err := kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tmpList := new(pb.ServerCfgInfo)
	tmpList.Index, _ = strconv.ParseUint(index, 10, 64)
	tmpList.ServerType = serverType
	for _, ev := range gresp.Kvs {
		tmpKey := key + "/local_address"
		if string(ev.Key) == tmpKey {
			tmpList.LocalAddress = string(ev.Value)
			continue
		}

		tmpKey = key + "/local_port"
		if string(ev.Key) == tmpKey {
			localPort, _ := strconv.Atoi(string(ev.Value))
			tmpList.LocalPort = uint32(localPort)
			continue
		}

		tmpKey = key + "/peer_address"
		if string(ev.Key) == tmpKey {
			tmpList.PeerAddress = string(ev.Value)
			continue
		}

		tmpKey = key + "/peer_port"
		if string(ev.Key) == tmpKey {
			peerPort, _ := strconv.Atoi(string(ev.Value))
			tmpList.PeerPort = uint32(peerPort)
		}

		tmpKey = key + "/monitor_cycle"
		if string(ev.Key) == tmpKey {
			tmpList.MonitorTime, _ = strconv.ParseUint(string(ev.Value), 10, 64)
			continue
		}

		tmpKey = key + "/max_update_times"
		if string(ev.Key) == tmpKey {
			maxUpdateTimes, _ := strconv.Atoi(string(ev.Value))
			tmpList.MaxUpdateTimes = uint32(maxUpdateTimes)
			continue
		}

		tmpKey = key + "/max_load_balancing_time"
		if string(ev.Key) == tmpKey {
			tmpList.MaxLoadBalancingTime, _ = strconv.ParseUint(string(ev.Value), 10, 64)
			continue
		}

		tmpKey = key + "/auto_update_lease_time"
		if string(ev.Key) == tmpKey {
			tmpList.AutoUpdateLeaseTime, _ = strconv.ParseUint(string(ev.Value), 10, 64)
			continue
		}

		tmpKey = key + "/separate_digit"
		if string(ev.Key) == tmpKey {
			separateDigit, _ := strconv.Atoi(string(ev.Value))
			tmpList.SeparateDigit = uint32(separateDigit)
			continue
		}
	}

	return tmpList, nil
}

func (ds *DhcpService) GetServerCfg(in *pb.GetServerCfgReq, stream pb.FailoverManager_GetServerCfgServer) error {
	logger.Debug("Enter GetServerCfg.")

	kvc := clientv3.NewKV(ds.Db)

	var serverInfo []*pb.ServerCfgInfo
	var serverType string

	prefix := ds.KeyPrefix + "address_manager/failover/"
	if in.ServerType == pb.ServerType_master_server {
		serverType = "master_server"
	} else if in.ServerType == pb.ServerType_slave_server {
		serverType = "slave_server"
	} else {
		return errors.New("The server type is Error.")
	}

	if serverType != "" && in.Index != 0 {
		prefix += serverType + "/"
		strIndex := strconv.FormatUint(in.Index, 10)
		prefix += "config_index/" + strIndex
		tmpServer, err := ds.queryServerInfo(kvc, in.ServerType, prefix, strIndex)
		if err == nil {
			serverInfo = append(serverInfo, tmpServer)
		}

	} else if serverType != "" && in.Index == 0 {
		commPrefix := ds.KeyPrefix + "address_manager/common/failover/"
		commPrefix += serverType + "/config_index/"
		var strKey, strIndex string
		gresp, _ := kvc.Get(context.TODO(), commPrefix, clientv3.WithPrefix())
		for _, ev := range gresp.Kvs {
			strKey = string(ev.Key)
			strIndex = strKey[len(commPrefix):]
			tmpKey := prefix
			tmpKey += serverType + "/"
			tmpKey += "config_index/" + strIndex
			tmpServer, err := ds.queryServerInfo(kvc, in.ServerType, tmpKey, strIndex)
			if err == nil {
				serverInfo = append(serverInfo, tmpServer)
			}
		}
	}

	var tmpStart, tmpEnd, number, remainder uint64
	number = uint64(len(serverInfo) / GRPC_RESULT_RATE)
	remainder = uint64(len(serverInfo) % GRPC_RESULT_RATE)

	var info []*pb.ServerCfgInfo
	var i uint64
	for i = 0; i < number; i++ {
		tmpStart = GRPC_RESULT_RATE * i
		tmpEnd = tmpStart + GRPC_RESULT_RATE
		info = serverInfo[tmpStart:tmpEnd]
		resp := &pb.GetServerCfgResp{
			RespServer: info,
		}
		stream.Send(resp)
	}

	if remainder != 0 {
		tmpStart = GRPC_RESULT_RATE * number
		tmpEnd = tmpStart + remainder
		info = serverInfo[tmpStart:tmpEnd]
		resp := &pb.GetServerCfgResp{
			RespServer: info,
		}
		stream.Send(resp)
	}

	logger.Debug("Exit.")
	return nil
}
