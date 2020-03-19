package dhcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"

	"reyzar.com/server-api/pkg/logger"

	"github.com/coreos/etcd/clientv3"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

func (ds *DhcpService) setOperGather(list []*pb.BlackAndWhiteListInfo, operType pb.OperationType) ([]*OperateGather, error) {
	commonPrefix := ds.KeyPrefix + "address_manager/common/access_control/"
	prefix := ds.KeyPrefix + "address_manager/access_control/"
	var listType, strIndex, tmpPrefix string
	oper := make([]*OperateGather, 0, 15)
	var tmpOper *OperateGather
	tmpOperType := int32(operType)

	for i := 0; i < len(list); i++ {
		if list[i].ListType == pb.ListType_black_list {
			listType = "black_list"
		} else if list[i].ListType == pb.ListType_white_list {
			listType = "white_list"
		} else {
			return nil, errors.New("The list type is Error.")
		}

		strIndex = strconv.FormatUint(list[i].Index, 10)
		tmpPrefix = prefix + listType + "/config_index/" + strIndex

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
			tmpOper.Key = commonPrefix + listType + "/config_index/" + strIndex
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
			continue
		}

		// mac address
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/mac"
		tmpOper.Value = list[i].MacAddress
		oper = append(oper, tmpOper)

		// divice name
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/user_name"
		tmpOper.Value = list[i].UserName
		oper = append(oper, tmpOper)

		//subnet
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/subnet_id"
		tmpOper.Value = strconv.FormatUint(list[i].SubnetId, 10)
		oper = append(oper, tmpOper)

		//notes
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/notes"
		tmpOper.Value = list[i].Notes
		oper = append(oper, tmpOper)

		//status
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/status"
		tmpOper.Value = fmt.Sprintf("%d", list[i].Status)
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + listType + "/config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (ds *DhcpService) SetBlackAndWhiteList(stream pb.AccessControl_SetBlackAndWhiteListServer) error {
	logger.Debug("Enter SetBlackAndWhiteList.")
	result := &pb.RespResult{}

	var oper []*OperateGather
	var tmpOper []*OperateGather
	var err error

	for {
		BWList, err := stream.Recv()
		if err == io.EOF {
			logger.Debug("All messages had received.")
			break
		}
		if err != nil {
			logger.Error(err)
			return err
		}

		tmpOper, err = ds.setOperGather(BWList.ListInfo, BWList.OperType)
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

func (ds *DhcpService) queryListInfo(kv clientv3.KV, listType pb.ListType, key, index string) (*pb.BlackAndWhiteListInfo, error) {
	gresp, err := kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tmpList := new(pb.BlackAndWhiteListInfo)
	tmpList.Index, _ = strconv.ParseUint(index, 10, 64)
	tmpList.ListType = listType
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

		tmpKey = key + "/subnet_id"
		if string(ev.Key) == tmpKey {
			tmpList.SubnetId, _ = strconv.ParseUint(string(ev.Value), 10, 64)
			continue
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

func (ds *DhcpService) GetBlackAndWhiteList(in *pb.GetBWListReq, stream pb.AccessControl_GetBlackAndWhiteListServer) error {
	logger.Debug("Enter GetBlackAndWhiteList.")

	kvc := clientv3.NewKV(ds.Db)

	var listInfo []*pb.BlackAndWhiteListInfo
	var listType string

	prefix := ds.KeyPrefix + "address_manager/access_control/"
	if in.ListType == pb.ListType_black_list {
		listType = "black_list"
	} else if in.ListType == pb.ListType_white_list {
		listType = "white_list"
	} else {
		return errors.New("The list type is Error.")
	}

	if listType != "" && in.Index != 0 {
		prefix += listType + "/"
		strIndex := strconv.FormatUint(in.Index, 10)
		prefix += "config_index/" + strIndex
		tmpList, err := ds.queryListInfo(kvc, in.ListType, prefix, strIndex)
		if err == nil {
			listInfo = append(listInfo, tmpList)
		}

	} else if listType != "" && in.Index == 0 {
		commPrefix := ds.KeyPrefix + "address_manager/common/access_control/"
		commPrefix += listType + "/config_index/"
		var strKey, strIndex string
		gresp, _ := kvc.Get(context.TODO(), commPrefix, clientv3.WithPrefix())
		for _, ev := range gresp.Kvs {
			strKey = string(ev.Key)
			strIndex = strKey[len(commPrefix):]
			tmpKey := prefix
			tmpKey += listType + "/"
			tmpKey += "config_index/" + strIndex
			tmpList, err := ds.queryListInfo(kvc, in.ListType, tmpKey, strIndex)
			log.Println(tmpList)
			if err == nil {
				listInfo = append(listInfo, tmpList)
			}
		}
	}

	var tmpStart, tmpEnd, number, remainder uint64
	number = uint64(len(listInfo) / GRPC_RESULT_RATE)
	remainder = uint64(len(listInfo) % GRPC_RESULT_RATE)

	var info []*pb.BlackAndWhiteListInfo
	var i uint64
	for i = 0; i < number; i++ {
		tmpStart = GRPC_RESULT_RATE * i
		tmpEnd = tmpStart + GRPC_RESULT_RATE
		info = listInfo[tmpStart:tmpEnd]
		resp := &pb.GetBWListResp{
			RespList: info,
		}
		stream.Send(resp)
	}

	if remainder != 0 {
		tmpStart = GRPC_RESULT_RATE * number
		tmpEnd = tmpStart + remainder
		info = listInfo[tmpStart:tmpEnd]
		resp := &pb.GetBWListResp{
			RespList: info,
		}
		stream.Send(resp)
	}

	logger.Debug("Exit.")
	return nil
}

func (ds *DhcpService) SetBWListStatus(ctx context.Context, in *pb.StatusBWListReq) (*pb.RespResult, error) {
	logger.Debug("Enter SetBWListStatus.")
	result := &pb.RespResult{}

	if in.Index <= 0 {
		result.ResultCode = ErrorList[1].ErrCode
		result.Description = ErrorList[1].ErrDesc
		return result, nil
	}

	var listType string
	if in.ListType == pb.ListType_black_list {
		listType = "black_list"
	} else if in.ListType == pb.ListType_white_list {
		listType = "white_list"
	} else {
		result.ResultCode = ErrorList[2].ErrCode
		result.Description = ErrorList[2].ErrDesc
		return result, nil
	}

	strIndex := strconv.FormatUint(in.Index, 10)

	strKey := ds.KeyPrefix + "address_manager/access_control/" +
		listType + "/config_index/" + strIndex + "/status"
	reply, err := ds.SetSingleKeyToDB(strKey, fmt.Sprintf("%d", in.Status))
	if err != nil {
		result.ResultCode = reply.ResultCode
		result.Description = reply.Description
	}

	logger.Debug("Exit.")
	return result, nil
}
