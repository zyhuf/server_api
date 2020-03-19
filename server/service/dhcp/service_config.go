package dhcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

func (ds *DhcpService) setSubnetCfgGather(cfg []*pb.SubnetInfo, operType pb.OperationType) ([]*OperateGather, error) {
	commonPrefix := ds.KeyPrefix + "address_manager/common/service_config/subnet_list/"
	prefix := ds.KeyPrefix + "address_manager/service_config/subnet_list/"

	var strIndex, tmpPrefix string
	oper := make([]*OperateGather, 0, 30)
	var tmpOper *OperateGather
	tmpOperType := int32(operType)

	for i := 0; i < len(cfg); i++ {
		strIndex = strconv.FormatUint(cfg[i].Index, 10)
		tmpPrefix = prefix + "config_index/" + strIndex

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
			tmpOper.Key = commonPrefix + "config_index/" + strIndex
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
			continue
		}

		//subnet name
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/subnet_name"
		tmpOper.Value = cfg[i].Name
		oper = append(oper, tmpOper)

		//vlan id
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/vlan_id"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].VlanId)
		oper = append(oper, tmpOper)

		//type
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/subnet_type"
		tmpOper.Value = fmt.Sprintf("%d", cfg[i].Type)
		oper = append(oper, tmpOper)

		//valid
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/subnet_valid"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].SubnetValid)
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/subnet_bw_type"
		tmpOper.Value = fmt.Sprintf("%d", cfg[i].BwType)
		oper = append(oper, tmpOper)

		//IPv4 subnet net
		if cfg[i].V4Sub != nil {
			for m := 0; m < len(cfg[i].V4Sub.NetSegment); m++ {
				segmentIndex := strconv.FormatUint(cfg[i].V4Sub.NetSegment[m].Index, 10)
				tmpOper = new(OperateGather)
				tmpOper.OperType = tmpOperType
				tmpOper.Key = tmpPrefix + "/IPv4_subnet/segment_index/" + segmentIndex + "/net"
				tmpOper.Value = cfg[i].V4Sub.NetSegment[m].Net
				oper = append(oper, tmpOper)
			}
			//Pool
			for k := 0; k < len(cfg[i].V4Sub.Pool); k++ {
				tmpOper = new(OperateGather)
				tmpOper.OperType = tmpOperType
				strPoolIndex := strconv.FormatUint(cfg[i].V4Sub.Pool[k].Index, 10)
				tmpOper.Key = tmpPrefix + "/IPv4_subnet/pool_index/" + strPoolIndex + "/addr_segment"
				tmpOper.Value = cfg[i].V4Sub.Pool[k].Beg + "-" + cfg[i].V4Sub.Pool[k].End
				oper = append(oper, tmpOper)
			}
		}

		//IPv6 subnet net
		if cfg[i].V6Sub != nil {
			for m := 0; m < len(cfg[i].V6Sub.NetSegment); m++ {
				segmentIndex := strconv.FormatUint(cfg[i].V6Sub.NetSegment[m].Index, 10)
				tmpOper = new(OperateGather)
				tmpOper.OperType = tmpOperType
				tmpOper.Key = tmpPrefix + "/IPv6_subnet/segment_index/" + segmentIndex + "/net"
				tmpOper.Value = cfg[i].V6Sub.NetSegment[m].Net
				oper = append(oper, tmpOper)
			}

			//Pool
			for k := 0; k < len(cfg[i].V6Sub.Pool); k++ {
				tmpOper = new(OperateGather)
				tmpOper.OperType = tmpOperType
				strPoolIndex := strconv.FormatUint(cfg[i].V6Sub.Pool[k].Index, 10)
				tmpOper.Key = tmpPrefix + "/IPv6_subnet/pool_index/" + strPoolIndex + "/addr_segment"
				tmpOper.Value = cfg[i].V6Sub.Pool[k].Beg + "-" + cfg[i].V6Sub.Pool[k].End
				oper = append(oper, tmpOper)
			}

			//PD Prefix
			tmpOper = new(OperateGather)
			tmpOper.OperType = tmpOperType
			tmpOper.Key = tmpPrefix + "/IPv6_subnet/pd_prefix_segment"
			if cfg[i].V6Sub.PdInfo != nil {
				tmpOper.Value = cfg[i].V6Sub.PdInfo.StartPrefix + "-" + cfg[i].V6Sub.PdInfo.EndPrefix
			}
			oper = append(oper, tmpOper)
		}

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + "config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (ds *DhcpService) OperateSubnet(ctx context.Context, in *pb.SubnetInfos) (*pb.RespResult, error) {
	logger.Debug("Enter OperateSubnet.")
	result := &pb.RespResult{}

	oper, _ := ds.setSubnetCfgGather(in.Ev, in.OperateType)

	if ds.OperateEtcdKv(oper) != nil {
		result.ResultCode = ErrorList[0].ErrCode
		result.Description = ErrorList[0].ErrDesc
	}

	logger.Debug("Exit.")
	return result, nil
}

func (ds *DhcpService) querySubnetInfo(kv clientv3.KV, key, index string) (*pb.SubnetInfo, error) {
	info := new(pb.SubnetInfo)
	info.Index, _ = strconv.ParseUint(index, 10, 64)

	key = key + index
	tmpKey := key + "/subnet_name"
	gresp, err := kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Name = string(gresp.Kvs[0].Value)
	}

	tmpKey = key + "/vlan_id"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		vlanID, _ := strconv.Atoi(string(gresp.Kvs[0].Value))
		info.VlanId = int32(vlanID)
	}

	tmpKey = key + "/subnet_type"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		tmpType, _ := strconv.Atoi(string(gresp.Kvs[0].Value))
		info.Type = pb.SubnetType(tmpType)
	}

	tmpKey = key + "/subnet_valid"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		tmpValid, _ := strconv.Atoi(string(gresp.Kvs[0].Value))
		info.SubnetValid = uint32(tmpValid)
	}

	tmpKey = key + "/subnet_bw_type"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		tmpBwType, _ := strconv.Atoi(string(gresp.Kvs[0].Value))
		info.BwType = pb.ListType(tmpBwType)
	}

	info.V4Sub = new(pb.V4SubInfo)
	tmpKey = key + "/IPv4_subnet/segment_index"
	gresp, err = kv.Get(context.TODO(), tmpKey, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	for i := 0; i < len(gresp.Kvs); i++ {
		segmentInfo := new(pb.NetworkSegmentInfo)
		strKey := strings.Split(string(gresp.Kvs[i].Key), "/")
		tmpIndex := strKey[len(strKey)-2]
		segmentInfo.Index, _ = strconv.ParseUint(tmpIndex, 10, 64)
		segmentInfo.Net = string(gresp.Kvs[i].Value)
		info.V4Sub.NetSegment = append(info.V4Sub.NetSegment, segmentInfo)
	}

	info.V6Sub = new(pb.V6SubInfo)
	tmpKey = key + "/IPv6_subnet/segment_index"
	gresp, err = kv.Get(context.TODO(), tmpKey, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	for i := 0; i < len(gresp.Kvs); i++ {
		segmentInfo := new(pb.NetworkSegmentInfo)
		strKey := strings.Split(string(gresp.Kvs[i].Key), "/")
		tmpIndex := strKey[len(strKey)-2]
		segmentInfo.Index, _ = strconv.ParseUint(tmpIndex, 10, 64)
		segmentInfo.Net = string(gresp.Kvs[i].Value)
		info.V6Sub.NetSegment = append(info.V6Sub.NetSegment, segmentInfo)
	}

	tmpKey = key + "/IPv4_subnet/pool_index/"
	gresp, err = kv.Get(context.TODO(), tmpKey, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	for i := 0; i < len(gresp.Kvs); i++ {
		poolInfo := new(pb.PoolInfo)
		strKey := strings.Split(string(gresp.Kvs[i].Key), "/")
		tmpIndex := strKey[len(strKey)-2]
		poolInfo.Index, _ = strconv.ParseUint(tmpIndex, 10, 64)
		addrSegment := strings.Split(string(gresp.Kvs[i].Value), "-")
		if len(addrSegment) == 2 {
			poolInfo.Beg = addrSegment[0]
			poolInfo.End = addrSegment[1]
		}
		info.V4Sub.Pool = append(info.V4Sub.Pool, poolInfo)
	}

	tmpKey = key + "/IPv6_subnet/pool_index/"
	gresp, err = kv.Get(context.TODO(), tmpKey, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	for i := 0; i < len(gresp.Kvs); i++ {
		poolInfo := new(pb.PoolInfo)
		strKey := strings.Split(string(gresp.Kvs[i].Key), "/")
		tmpIndex := strKey[len(strKey)-2]
		poolInfo.Index, _ = strconv.ParseUint(tmpIndex, 10, 64)
		addrSegment := strings.Split(string(gresp.Kvs[i].Value), "-")
		if len(addrSegment) == 2 {
			poolInfo.Beg = addrSegment[0]
			poolInfo.End = addrSegment[1]
		}
		info.V6Sub.Pool = append(info.V6Sub.Pool, poolInfo)
	}

	tmpKey = key + "/IPv6_subnet/pd_prefix_segment"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		tmpPrefix := strings.Split(string(gresp.Kvs[0].Value), "-")
		if len(tmpPrefix) == 2 {
			pdPrefix := new(pb.PDPrefixInfo)
			pdPrefix.StartPrefix = tmpPrefix[0]
			pdPrefix.EndPrefix = tmpPrefix[1]
			info.V6Sub.PdInfo = pdPrefix
		}
	}

	return info, nil

}

func (ds *DhcpService) GetSubnet(ctx context.Context, in *pb.ReqStatus) (*pb.SubnetInfos, error) {
	logger.Debug("Enter GetSubnet.")

	kvc := clientv3.NewKV(ds.Db)

	prefix := ds.KeyPrefix + "address_manager/service_config/subnet_list/config_index/"
	strIndex := strconv.FormatUint(in.Index, 10)

	info, err := ds.querySubnetInfo(kvc, prefix, strIndex)
	if err != nil {
		return nil, err
	}

	var subnetInfo []*pb.SubnetInfo
	subnetInfo = append(subnetInfo, info)
	resp := &pb.SubnetInfos{
		Ev: subnetInfo,
	}

	logger.Debug("Exit.")
	return resp, nil
}

func (ds *DhcpService) setStaticGather(cfg []*pb.IpStaticInfo, operType pb.OperationType) ([]*OperateGather, error) {
	commonPrefix := ds.KeyPrefix + "address_manager/common/service_config/static_bound/"
	prefix := ds.KeyPrefix + "address_manager/service_config/static_bound/"

	var strIndex, tmpPrefix string
	oper := make([]*OperateGather, 0, 15)
	var tmpOper *OperateGather
	tmpOperType := int32(operType)

	for i := 0; i < len(cfg); i++ {
		strIndex = strconv.FormatUint(cfg[i].Index, 10)
		tmpPrefix = prefix + "config_index/" + strIndex

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
			tmpOper.Key = commonPrefix + "config_index/" + strIndex
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
			continue
		}

		//subnet id
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/subnet_id"
		tmpOper.Value = strconv.FormatUint(cfg[i].SubnetId, 10)
		oper = append(oper, tmpOper)

		//mac addr
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/mac"
		tmpOper.Value = cfg[i].Mac
		oper = append(oper, tmpOper)

		//duid
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/duid"
		tmpOper.Value = cfg[i].Duid
		oper = append(oper, tmpOper)

		//IPv4 addr
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/IPv4_addr"
		tmpOper.Value = cfg[i].V4Ip
		oper = append(oper, tmpOper)

		//IPv6 addr
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/IPv6_addr"
		tmpOper.Value = cfg[i].V6Ip
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + "config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (ds *DhcpService) OperateStatic(ctx context.Context, in *pb.IpStaticInfos) (*pb.RespResult, error) {
	logger.Debug("Enter OperateStatic.")
	result := &pb.RespResult{}

	oper, _ := ds.setStaticGather(in.Ev, in.OperateType)
	if ds.OperateEtcdKv(oper) != nil {
		result.ResultCode = ErrorList[0].ErrCode
		result.Description = ErrorList[0].ErrDesc
	}

	logger.Debug("Exit.")
	return result, nil
}

func (ds *DhcpService) queryStaticInfo(kv clientv3.KV, key, index string) (*pb.IpStaticInfo, error) {
	info := new(pb.IpStaticInfo)
	info.Index, _ = strconv.ParseUint(index, 10, 64)
	key += index

	tmpKey := key + "/subnet_id"
	gresp, err := kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.SubnetId, _ = strconv.ParseUint(string(gresp.Kvs[0].Value), 10, 64)
	}

	tmpKey = key + "/mac"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Mac = string(gresp.Kvs[0].Value)
	}

	tmpKey = key + "/duid"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Duid = string(gresp.Kvs[0].Value)
	}

	tmpKey = key + "/IPv4_addr"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.V4Ip = string(gresp.Kvs[0].Value)
	}

	tmpKey = key + "/IPv6_addr"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.V6Ip = string(gresp.Kvs[0].Value)
	}

	return info, nil
}

func (ds *DhcpService) GetStatic(ctx context.Context, in *pb.ReqStatus) (*pb.IpStaticInfos, error) {
	logger.Debug("Enter GetStatic.")
	kvc := clientv3.NewKV(ds.Db)

	prefix := ds.KeyPrefix + "address_manager/service_config/static_bound/config_index/"
	strIndex := strconv.FormatUint(in.Index, 10)

	info, err := ds.queryStaticInfo(kvc, prefix, strIndex)
	if err != nil {
		return nil, err
	}

	var staticInfo []*pb.IpStaticInfo
	staticInfo = append(staticInfo, info)
	resp := &pb.IpStaticInfos{
		Ev: staticInfo,
	}

	logger.Debug("Exit.")
	return resp, nil
}
