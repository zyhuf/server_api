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

func (ds *DhcpService) setAuthGather(cfg []*pb.AuthCfgInfo, operType pb.OperationType) ([]*OperateGather, error) {
	commonPrefix := ds.KeyPrefix + "address_manager/common/auth/"
	prefix := ds.KeyPrefix + "address_manager/auth/"

	var authType, strIndex, tmpPrefix string
	oper := make([]*OperateGather, 0, 15)
	var tmpOper *OperateGather
	tmpOperType := int32(operType)

	for i := 0; i < len(cfg); i++ {
		if cfg[i].Type == pb.AuthType_perpetual_auth {
			authType = "perpetual_auth"
		} else if cfg[i].Type == pb.AuthType_temporary_auth {
			authType = "temporary_auth"
		} else {
			return nil, errors.New("The auth type is Error.")
		}

		strIndex = strconv.FormatUint(cfg[i].Index, 10)
		tmpPrefix = prefix + authType + "/config_index/" + strIndex

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
			tmpOper.Key = commonPrefix + authType + "/config_index/" + strIndex
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
			continue
		}

		//mac address
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/mac"
		tmpOper.Value = cfg[i].MacAddress
		oper = append(oper, tmpOper)

		//user name
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/user_name"
		tmpOper.Value = cfg[i].UserName
		oper = append(oper, tmpOper)

		//expire time
		if cfg[i].Type == pb.AuthType_temporary_auth {
			tmpOper = new(OperateGather)
			tmpOper.OperType = tmpOperType
			tmpOper.Key = tmpPrefix + "/expire_time"
			tmpOper.Value = fmt.Sprintf("%v", cfg[i].ExpireTime)
			oper = append(oper, tmpOper)

			tmpOper = new(OperateGather)
			tmpOper.OperType = tmpOperType
			tmpOper.Key = tmpPrefix + "/expire_date"
			tmpOper.Value = cfg[i].ExpireDate
			oper = append(oper, tmpOper)
		}

		//notes
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/notes"
		tmpOper.Value = cfg[i].Notes
		oper = append(oper, tmpOper)

		//status
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/status"
		tmpOper.Value = fmt.Sprintf("%d", cfg[i].Status)
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + authType + "/config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (ds *DhcpService) SetAuthCfg(stream pb.AuthManager_SetAuthCfgServer) error {
	logger.Debug("Enter SetAuthCfg.")
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

		tmpOper, err = ds.setAuthGather(cfg.AuthConfig, cfg.OperType)
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

func (ds *DhcpService) queryAuthInfo(kv clientv3.KV, authType pb.AuthType, key, index string) (*pb.AuthCfgInfo, error) {
	gresp, err := kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tmpList := new(pb.AuthCfgInfo)
	tmpList.Index, _ = strconv.ParseUint(index, 10, 64)
	tmpList.Type = authType
	for _, ev := range gresp.Kvs {
		tmpKey := key + "/mac"
		if string(ev.Key) == tmpKey {
			tmpList.MacAddress = string(ev.Value)
			continue
		}

		tmpKey = key + "/user_name"
		if string(ev.Key) == tmpKey {
			tmpList.UserName = string(ev.Value)
			continue
		}

		if tmpList.Type == pb.AuthType_temporary_auth {
			tmpKey = key + "/expire_time"
			if string(ev.Key) == tmpKey {
				expireTime, _ := strconv.Atoi(string(ev.Value))
				tmpList.ExpireTime = uint32(expireTime)
				continue
			}

			tmpKey = key + "/expire_date"
			if string(ev.Key) == tmpKey {
				tmpList.ExpireDate = string(ev.Value)
				continue
			}
		}

		tmpKey = key + "/notes"
		if string(ev.Key) == tmpKey {
			tmpList.Notes = string(ev.Value)
			continue
		}

		tmpKey = key + "/status"
		if string(ev.Key) == tmpKey {
			status, _ := strconv.Atoi(string(ev.Value))
			tmpList.Status = pb.Status(status)
			continue
		}
	}

	return tmpList, nil
}

func (ds *DhcpService) GetAuthCfg(in *pb.GetAuthCfgReq, stream pb.AuthManager_GetAuthCfgServer) error {
	logger.Debug("Enter GetAuthCfg.")
	kvc := clientv3.NewKV(ds.Db)

	var authInfo []*pb.AuthCfgInfo
	var authType string

	prefix := ds.KeyPrefix + "address_manager/auth/"
	if in.Type == pb.AuthType_perpetual_auth {
		authType = "perpetual_auth"
	} else if in.Type == pb.AuthType_temporary_auth {
		authType = "temporary_auth"
	} else {
		return errors.New("The auth type is Error.")
	}

	if authType != "" && in.Index != 0 {
		prefix += authType + "/"
		strIndex := strconv.FormatUint(in.Index, 10)
		prefix += "config_index/" + strIndex
		tmpAuth, err := ds.queryAuthInfo(kvc, in.Type, prefix, strIndex)
		if err == nil {
			authInfo = append(authInfo, tmpAuth)
		}

	} else if authType != "" && in.Index == 0 {
		commPrefix := ds.KeyPrefix + "address_manager/common/auth/"
		commPrefix += authType + "/config_index/"
		var strKey, strIndex string
		gresp, _ := kvc.Get(context.TODO(), commPrefix, clientv3.WithPrefix())
		for _, ev := range gresp.Kvs {
			strKey = string(ev.Key)
			strIndex = strKey[len(commPrefix):]
			tmpKey := prefix
			tmpKey += authType + "/"
			tmpKey += "config_index/" + strIndex
			tmpServer, err := ds.queryAuthInfo(kvc, in.Type, tmpKey, strIndex)
			if err == nil {
				authInfo = append(authInfo, tmpServer)
			}
		}
	}

	var tmpStart, tmpEnd, number, remainder uint64
	number = uint64(len(authInfo) / GRPC_RESULT_RATE)
	remainder = uint64(len(authInfo) % GRPC_RESULT_RATE)

	var info []*pb.AuthCfgInfo
	var i uint64
	for i = 0; i < number; i++ {
		tmpStart = GRPC_RESULT_RATE * i
		tmpEnd = tmpStart + GRPC_RESULT_RATE
		info = authInfo[tmpStart:tmpEnd]
		resp := &pb.GetAuthCfgResp{
			RespAuth: info,
		}
		stream.Send(resp)
	}

	if remainder != 0 {
		tmpStart = GRPC_RESULT_RATE * number
		tmpEnd = tmpStart + remainder
		info = authInfo[tmpStart:tmpEnd]
		resp := &pb.GetAuthCfgResp{
			RespAuth: info,
		}
		stream.Send(resp)
	}

	logger.Debug("Exit.")
	return nil
}

func (ds *DhcpService) SetAuthStatus(ctx context.Context, in *pb.StatusAuthReq) (*pb.RespResult, error) {
	logger.Debug("Enter SetAuthStatus.")
	result := &pb.RespResult{}

	if in.Index <= 0 {
		result.ResultCode = ErrorList[1].ErrCode
		result.Description = ErrorList[1].ErrDesc
		return result, nil
	}

	var authType string
	if in.Type == pb.AuthType_perpetual_auth {
		authType = "perpetual_auth"
	} else if in.Type == pb.AuthType_temporary_auth {
		authType = "temporary_auth"
	} else {
		result.ResultCode = ErrorList[3].ErrCode
		result.Description = ErrorList[3].ErrDesc
		return result, nil
	}

	strIndex := strconv.FormatUint(in.Index, 10)
	strKey := ds.KeyPrefix + "address_manager/auth/" +
		authType + "/config_index/" + strIndex + "/status"
	reply, err := ds.SetSingleKeyToDB(strKey, fmt.Sprintf("%d", in.Status))
	if err != nil {
		result.ResultCode = reply.ResultCode
		result.Description = reply.Description
	}

	logger.Debug("Exit.")
	return result, nil
}

func (ds *DhcpService) setDisableListGather(cfg []*pb.DisableUserListInfo, operType pb.OperationType) ([]*OperateGather, error) {
	commonPrefix := ds.KeyPrefix + "address_manager/common/auth/disable_list"
	prefix := ds.KeyPrefix + "address_manager/auth/disable_list"

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

		//MacAddress
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/mac"
		tmpOper.Value = cfg[i].MacAddress
		oper = append(oper, tmpOper)

		//user name
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/user_name"
		tmpOper.Value = cfg[i].UserName
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + "/config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (ds *DhcpService) SetDisableUserList(stream pb.AuthManager_SetDisableUserListServer) error {
	logger.Debug("Enter SetDisableUserList.")
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

		tmpOper, err = ds.setDisableListGather(cfg.CfgInfo, cfg.OperType)
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

func (ds *DhcpService) queryDisableListInfo(kv clientv3.KV, key, index string) (*pb.DisableUserListInfo, error) {
	gresp, err := kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tmpList := new(pb.DisableUserListInfo)
	tmpList.Index, _ = strconv.ParseUint(index, 10, 64)

	for _, ev := range gresp.Kvs {
		tmpKey := key + "/mac"
		if string(ev.Key) == tmpKey {
			tmpList.MacAddress = string(ev.Value)
			continue
		}

		tmpKey = key + "/user_name"
		if string(ev.Key) == tmpKey {
			tmpList.UserName = string(ev.Value)
			continue
		}
	}

	return tmpList, nil
}

func (ds *DhcpService) GetDisableUserList(in *pb.ReqStatus, stream pb.AuthManager_GetDisableUserListServer) error {
	logger.Debug("Enter GetDisableUserList.")
	if in == nil {
		return errors.New("Input Parameter is error.")
	}

	var disableList []*pb.DisableUserListInfo
	kvc := clientv3.NewKV(ds.Db)
	prefix := ds.KeyPrefix + "address_manager/auth/disable_list/"
	if in.Index != 0 {
		strIndex := strconv.FormatUint(in.Index, 10)
		prefix += "config_index/" + strIndex
		tmpList, err := ds.queryDisableListInfo(kvc, prefix, strIndex)
		if err == nil {
			disableList = append(disableList, tmpList)
		}

	} else if in.Index == 0 {
		commPrefix := ds.KeyPrefix + "address_manager/common/auth/disable_list/config_index/"
		var strKey, strIndex string
		gresp, _ := kvc.Get(context.TODO(), commPrefix, clientv3.WithPrefix())
		for _, ev := range gresp.Kvs {
			strKey = string(ev.Key)
			strIndex = strKey[len(commPrefix):]
			tmpKey := prefix
			tmpKey += "config_index/" + strIndex
			tmpList, err := ds.queryDisableListInfo(kvc, tmpKey, strIndex)
			if err == nil {
				disableList = append(disableList, tmpList)
			}
		}
	}

	var tmpStart, tmpEnd, number, remainder uint64
	number = uint64(len(disableList) / GRPC_RESULT_RATE)
	remainder = uint64(len(disableList) % GRPC_RESULT_RATE)

	var info []*pb.DisableUserListInfo
	var i uint64
	for i = 0; i < number; i++ {
		tmpStart = GRPC_RESULT_RATE * i
		tmpEnd = tmpStart + GRPC_RESULT_RATE
		info = disableList[tmpStart:tmpEnd]
		resp := &pb.DisableUserListRsp{
			CfgInfo: info,
		}
		stream.Send(resp)
	}

	if remainder != 0 {
		tmpStart = GRPC_RESULT_RATE * number
		tmpEnd = tmpStart + remainder
		info = disableList[tmpStart:tmpEnd]
		resp := &pb.DisableUserListRsp{
			CfgInfo: info,
		}
		stream.Send(resp)
	}

	logger.Debug("Exit.")
	return nil
}
