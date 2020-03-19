package dns

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"reyzar.com/server-api/pkg/dig"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dnsserver"
)

type ForwardResolver struct {
	DomainId     uint64
	ForwardId    uint64
	DomainName   string
	ResolverType uint16
	ResolverVal  string
	Status       int32 //0: success -1: abortnormal
}

func (d *DnsService) setForwardCfgGather(cfg []*pb.ForwardInfo, operType pb.OperType) ([]*OperateGather, error) {
	commonPrefix := d.KeyPrefix + "domain_manager/common/domain_config/"
	prefix := d.KeyPrefix + "domain_manager/domain_config/"

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

		//noets
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/notes"
		tmpOper.Value = cfg[i].Reference
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + "config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (d *DnsService) OperateForward(ctx context.Context, req *pb.ForwardInfos) (*pb.RespStatus, error) {
	logger.Debug("Enter OperateForward.")
	result := &pb.RespStatus{}

	if req == nil {
		result.Code = ErrorList[2].ErrCode
		result.Msg = ErrorList[2].ErrDesc
		return result, nil
	}

	oper, _ := d.setForwardCfgGather(req.Ev, req.OperateType)

	if d.OperateEtcdKv(oper) != nil {
		result.Code = ErrorList[0].ErrCode
		result.Msg = ErrorList[0].ErrDesc
	}

	logger.Debug("Exit.")

	return result, nil
}

func (ds *DnsService) queryForwardInfo(kv clientv3.KV, key, index string) (*pb.ForwardInfo, error) {
	info := new(pb.ForwardInfo)
	info.Id, _ = strconv.ParseUint(index, 10, 64)
	key += index

	// domain name
	tmpKey := key + "/domain_name"
	gresp, err := kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Domain = string(gresp.Kvs[0].Value)
	}

	// notes
	tmpKey = key + "/notes"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Reference = string(gresp.Kvs[0].Value)
	}

	return info, nil
}

func (d *DnsService) queryAllId(kv clientv3.KV, prefix string) []uint64 {
	var id []uint64
	resp, err := kv.Get(context.TODO(), prefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return id
	}

	var strKey string
	for _, ev := range resp.Kvs {
		strKey = string(ev.Key)
		index := strKey[len(prefix):]
		tmpId, _ := strconv.ParseUint(index, 10, 64)
		id = append(id, tmpId)
	}

	return id
}

func (d *DnsService) GetForward(ctx context.Context, req *pb.ReqStatus) (*pb.ForwardInfos, error) {
	logger.Debug("Enter GetForward.")

	if req == nil {
		logger.Debug("Input Parameter ReqStatus is nil.")
		return nil, nil
	}

	kvc := clientv3.NewKV(d.Db)
	var index []uint64
	if len(req.Id) > 0 && req.Id[0] == 0 {
		prefix := d.KeyPrefix + "domain_manager/common/domain_config/config_index"
		index = d.queryAllId(kvc, prefix)
	} else {
		index = append(index, req.Id...)
	}

	prefix := d.KeyPrefix + "domain_manager/domain_config/config_index/"

	var forwardInfo []*pb.ForwardInfo
	for i := 0; i < len(index); i++ {
		strIndex := strconv.FormatUint(index[i], 10)
		info, err := d.queryForwardInfo(kvc, prefix, strIndex)
		if err != nil {
			return nil, err
		}
		forwardInfo = append(forwardInfo, info)
	}

	resp := &pb.ForwardInfos{
		Ev: forwardInfo,
	}

	return resp, nil
}

func (d *DnsService) setDomainCfgGather(cfg []*pb.DomainInfo, operType pb.OperType) ([]*OperateGather, error) {
	commonPrefix := d.KeyPrefix + "domain_manager/common/forward_resolver/"
	prefix := d.KeyPrefix + "domain_manager/forward_resolver/"

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
			if cfg[i].ForId != 0 {
				tmpOper.Key = tmpOper.Key + "/domain_config_id/" + strconv.FormatUint(cfg[i].ForId, 10)
			}
			tmpOper.Value = ""
			oper = append(oper, tmpOper)

			tmpOper = new(OperateGather)
			tmpOper.OperType = tmpOperType
			tmpOper.Key = commonPrefix + "config_index/" + strIndex
			tmpOper.Value = ""
			oper = append(oper, tmpOper)
			continue
		}

		//domain_config_id
		tmpPrefix = tmpPrefix + "/domain_config_id/" + strconv.FormatUint(cfg[i].ForId, 10)

		//name
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/name"
		tmpOper.Value = cfg[i].Name
		oper = append(oper, tmpOper)

		//type
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/type"
		tmpOper.Value = cfg[i].Type
		oper = append(oper, tmpOper)

		//link
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/link"
		tmpOper.Value = strconv.FormatUint(cfg[i].IspId, 10)
		oper = append(oper, tmpOper)

		//value
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/value"
		tmpOper.Value = cfg[i].Value
		oper = append(oper, tmpOper)

		//TTL
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/ttl"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].Ttl)
		oper = append(oper, tmpOper)

		//Mx
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/mx_priority"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].Mx)
		oper = append(oper, tmpOper)

		//status
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = tmpPrefix + "/status"
		tmpOper.Value = fmt.Sprintf("%v", cfg[i].Enable)
		oper = append(oper, tmpOper)

		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = commonPrefix + "config_index/" + strIndex
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (d *DnsService) OperateDomain(ctx context.Context, req *pb.DomainInfos) (*pb.RespStatus, error) {
	logger.Debug("Enter OperateDomain.")

	result := &pb.RespStatus{}
	logger.Debug("req:", req)

	if req == nil {
		result.Code = ErrorList[2].ErrCode
		result.Msg = ErrorList[2].ErrDesc
		return result, nil
	}

	oper, _ := d.setDomainCfgGather(req.Ev, req.OperateType)

	if d.OperateEtcdKv(oper) != nil {
		result.Code = ErrorList[0].ErrCode
		result.Msg = ErrorList[0].ErrDesc
	}

	logger.Debug("Exit.")
	return result, nil
}

func (ds *DnsService) queryDomainInfo(kv clientv3.KV, key, index, forwardId string) (*pb.DomainInfo, error) {
	info := new(pb.DomainInfo)
	info.Id, _ = strconv.ParseUint(index, 10, 64)
	key += index

	// domain_config_id
	info.ForId, _ = strconv.ParseUint(forwardId, 10, 64)
	key += "/domain_config_id/" + forwardId

	//name
	tmpKey := key + "/name"
	log.Println(tmpKey)
	gresp, err := kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Name = string(gresp.Kvs[0].Value)
	}

	//type
	tmpKey = key + "/type"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Type = string(gresp.Kvs[0].Value)
	}

	//link
	tmpKey = key + "/link"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.IspId, _ = strconv.ParseUint(string(gresp.Kvs[0].Value), 10, 64)
	}

	//value
	tmpKey = key + "/value"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		info.Value = string(gresp.Kvs[0].Value)
	}

	//TTL
	tmpKey = key + "/ttl"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		tmpTTL, _ := strconv.Atoi(string(gresp.Kvs[0].Value))
		info.Ttl = int32(tmpTTL)
	}

	//mx_priority
	tmpKey = key + "/mx_priority"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		tmpMx, _ := strconv.Atoi(string(gresp.Kvs[0].Value))
		info.Mx = int32(tmpMx)
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

func (d *DnsService) GetDomain(ctx context.Context, req *pb.ForwardRefer) (*pb.DomainInfos, error) {
	logger.Debug("Enter GetDomain.")

	if req == nil {
		logger.Debug("Input Parameter ReqStatus is nil.")
		return nil, nil
	}

	kvc := clientv3.NewKV(d.Db)

	prefix := d.KeyPrefix + "domain_manager/forward_resolver/config_index/"

	forwardId := strconv.FormatUint(req.ForwardId, 10)
	index := strconv.FormatUint(req.DomainId, 10)

	info, err := d.queryDomainInfo(kvc, prefix, index, forwardId)
	if err != nil {
		return nil, err
	}

	var domainInfo []*pb.DomainInfo
	domainInfo = append(domainInfo, info)

	resp := &pb.DomainInfos{
		Ev: domainInfo,
	}

	return resp, nil

}

func (d *DnsService) EnableDomain(ctx context.Context, req *pb.ForwardEnableMsg) (*pb.RespStatus, error) {
	logger.Debug("Enter EnableDomain.")

	result := &pb.RespStatus{}

	if req == nil {
		result.Code = ErrorList[2].ErrCode
		result.Msg = ErrorList[2].ErrDesc
		return result, nil
	}

	strIndex := strconv.FormatUint(req.Refer.DomainId, 10)
	forwardId := strconv.FormatUint(req.Refer.ForwardId, 10)

	strKey := d.KeyPrefix + "domain_manager/forward_resolver/config_index/" +
		strIndex + "/domain_config_id/" + forwardId + "/status"
	log.Println(strKey)
	reply, err := d.SetSingleKeyToDB(strKey, fmt.Sprintf("%v", req.Enable))
	if err != nil {
		result.Code = reply.Code
		result.Msg = reply.Msg
	}

	logger.Debug("Exit.")
	return result, nil
}

func ResolverDomain(resolver []*ForwardResolver, addr []string, port string) {
	for i := 0; i < len(addr); i++ {
		var DnsDig dig.Dig
		DnsDig.Port = port
		DnsDig.SetDNS(addr[i])

		for j := 0; j < len(resolver); j++ {
			if resolver[j].Status == 0 {
				continue
			}

			msg, err := DnsDig.GetMsg(resolver[j].ResolverType, resolver[j].DomainName)
			if err != nil {
				resolver[j].Status = -1
				continue
			}

			for k := 0; k < len(msg.Answer); k++ {
				if strings.Contains(msg.Answer[k].String(), resolver[j].ResolverVal) {
					resolver[j].Status = 0
				}
			}
		}
	}
}

func (d *DnsService) GetServiceAddr(kvc clientv3.KV) []string {
	var addr []string
	prefix := d.KeyPrefix + "domain_manager/service_config/service_ip/"
	gresp, err := kvc.Get(context.TODO(), prefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return addr
	}

	for _, ev := range gresp.Kvs {
		strKey := string(ev.Key)
		tmpAddr := strKey[len(prefix):]
		addr = append(addr, tmpAddr)
	}

	return addr
}

func (d *DnsService) GetServicePort(kvc clientv3.KV) string {
	var port string
	strKey := d.KeyPrefix + "domain_manager/service_config/service_port"
	gresp, err := kvc.Get(context.TODO(), strKey)
	if err != nil {
		logger.Error(err)
		return port
	}

	if len(gresp.Kvs) != 0 {
		port = string(gresp.Kvs[0].Value)
	}

	return port
}

func MapResolverType(resolverType string) uint16 {
	var tmpType uint16
	if resolverType == "A" {
		tmpType = 1
	} else if resolverType == "AAAA" {
		tmpType = 28
	} else if resolverType == "NS" {
		tmpType = 43
	} else if resolverType == "MX" {
		tmpType = 15
	} else if resolverType == "SRV" {
		tmpType = 33
	} else if resolverType == "TXT" {
		tmpType = 16
	} else if resolverType == "CNAME" {
		tmpType = 5
	} else if resolverType == "CAA" {
		tmpType = 257
	}

	return tmpType
}

func (d *DnsService) GetResolverInfo(kvc clientv3.KV, ev []*pb.ForStatusInfo) []*ForwardResolver {
	var resolver []*ForwardResolver
	domainPrefix := d.KeyPrefix + "domain_manager/domain_config/config_index/"
	fowardPrefix := d.KeyPrefix + "domain_manager/forward_resolver/config_index/"
	for i := 0; i < len(ev); i++ {
		tmpResolver := new(ForwardResolver)
		tmpResolver.Status = -1
		tmpResolver.DomainId = ev[i].Refer.DomainId
		tmpResolver.ForwardId = ev[i].Refer.ForwardId
		domanIndex := strconv.FormatUint(ev[i].Refer.DomainId, 10)
		tmpKey := domainPrefix + domanIndex + "/domain_name"

		gresp, err := kvc.Get(context.TODO(), tmpKey)
		if err != nil {
			logger.Error(err)
			return resolver
		}
		if len(gresp.Kvs) != 0 {
			tmpResolver.DomainName = string(gresp.Kvs[0].Value)
		}

		forwardIndex := strconv.FormatUint(ev[i].Refer.ForwardId, 10)
		tmpKey = fowardPrefix + forwardIndex + "/domain_config_id/" + domanIndex + "/type"
		gresp, err = kvc.Get(context.TODO(), tmpKey)
		if err != nil {
			logger.Error(err)
			return resolver
		}
		if len(gresp.Kvs) != 0 {
			tmpResolver.ResolverType = MapResolverType(string(gresp.Kvs[0].Value))
		}

		tmpKey = fowardPrefix + forwardIndex + "/domain_config_id/" + domanIndex + "/value"
		gresp, err = kvc.Get(context.TODO(), tmpKey)
		if err != nil {
			logger.Error(err)
			return resolver
		}
		if len(gresp.Kvs) != 0 {
			tmpResolver.ResolverVal = string(gresp.Kvs[0].Value)
		}

		resolver = append(resolver, tmpResolver)
	}

	return resolver
}

func (d *DnsService) GetForwardStatus(ctx context.Context, req *pb.ForwardStatusInfos) (*pb.ForwardStatusInfos, error) {
	logger.Debug("Enter GetForwardStatus.")

	if req == nil {
		logger.Error("Input Parameter ForwardStatusInfos is nil.")
		return nil, nil
	}

	kvc := clientv3.NewKV(d.Db)

	resolver := d.GetResolverInfo(kvc, req.Ev)
	address := d.GetServiceAddr(kvc)
	servicePort := d.GetServicePort(kvc)

	ResolverDomain(resolver, address, servicePort)

	var info []*pb.ForStatusInfo
	for index := 0; index < len(resolver); index++ {
		refer := new(pb.ForwardRefer)
		refer.DomainId = resolver[index].DomainId
		refer.ForwardId = resolver[index].ForwardId

		tmpInfo := new(pb.ForStatusInfo)
		tmpInfo.Refer = refer
		tmpInfo.Status = resolver[index].Status

		info = append(info, tmpInfo)
	}

	rsp := &pb.ForwardStatusInfos{
		Ev: info,
	}

	return rsp, nil
}
