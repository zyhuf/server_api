package dns

import (
	"context"
	"fmt"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dnsserver"
)

func (d *DnsService) setSrvCfgGather(cfg *pb.SysConfs) ([]*OperateGather, error) {
	prefix := d.KeyPrefix + "domain_manager/service_config"

	oper := make([]*OperateGather, 0, 30)
	var tmpOper *OperateGather
	tmpOperType := int32(pb.OperType_ADD)

	if cfg.Sys == nil {
		logger.Error("Sys pointer is nil.")
		return nil, nil
	}
	//service port
	tmpOper = new(OperateGather)
	tmpOper.OperType = tmpOperType
	tmpOper.Key = prefix + "/service_port"
	tmpOper.Value = fmt.Sprintf("%d", cfg.Sys.Port)
	oper = append(oper, tmpOper)

	//cname_priority_enable
	tmpOper = new(OperateGather)
	tmpOper.OperType = tmpOperType
	tmpOper.Key = prefix + "/cname_priority_enable"
	tmpOper.Value = fmt.Sprintf("%v", cfg.Sys.CnamePriority)
	oper = append(oper, tmpOper)

	//tcp_enable
	tmpOper = new(OperateGather)
	tmpOper.OperType = tmpOperType
	tmpOper.Key = prefix + "/tcp_enable"
	tmpOper.Value = fmt.Sprintf("%v", cfg.Sys.TcpEnable)
	oper = append(oper, tmpOper)

	for i := 0; i < len(cfg.Ip); i++ {
		//IP
		tmpOper = new(OperateGather)
		tmpOper.OperType = tmpOperType
		tmpOper.Key = prefix + "/service_ip/" + cfg.Ip[i].Ip
		tmpOper.Value = ""
		oper = append(oper, tmpOper)
	}

	return oper, nil
}

func (d *DnsService) UpdateSysConf(ctx context.Context, req *pb.SysConfs) (*pb.RespStatus, error) {
	logger.Debug("Enter UpdateSysConf.")

	result := &pb.RespStatus{}

	if req == nil {
		result.Code = ErrorList[2].ErrCode
		result.Msg = ErrorList[2].ErrDesc
		return result, nil
	}

	rsp, _ := d.UpdateRecursion(ctx, req.Info)
	if rsp.Code != 0 {
		result.Code = rsp.Code
		result.Msg = rsp.Msg
		return result, nil
	}

	oper, _ := d.setSrvCfgGather(req)

	if d.OperateEtcdKv(oper) != nil {
		result.Code = ErrorList[0].ErrCode
		result.Msg = ErrorList[0].ErrDesc
	}

	logger.Debug("Exit.")
	return result, nil
}

func (d *DnsService) GetSysConf(ctx context.Context, req *pb.ReqStatus) (*pb.SysConfs, error) {
	logger.Debug("Enter GetSysConf.")

	key := d.KeyPrefix + "domain_manager/service_config"

	kv := clientv3.NewKV(d.Db)
	conf := new(pb.SysConf)
	// service port
	tmpKey := key + "/service_port"
	gresp, err := kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		tmpPort, _ := strconv.Atoi(string(gresp.Kvs[0].Value))
		conf.Port = int32(tmpPort)
	}

	//cname_priority_enable
	tmpKey = key + "/cname_priority_enable"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		conf.CnamePriority, _ = strconv.ParseBool(string(gresp.Kvs[0].Value))
	}

	//tcp_enable
	tmpKey = key + "/tcp_enable"
	gresp, err = kv.Get(context.TODO(), tmpKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if len(gresp.Kvs) != 0 {
		conf.TcpEnable, _ = strconv.ParseBool(string(gresp.Kvs[0].Value))
	}

	//service ip
	tmpKey = key + "/service_ip"
	gresp, err = kv.Get(context.TODO(), tmpKey, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var serviceIp []*pb.ServiceIPInfo
	var strKey string
	for _, ev := range gresp.Kvs {
		strKey = string(ev.Key)
		strIP := strKey[len(tmpKey)+1:]
		srvinfo := new(pb.ServiceIPInfo)
		srvinfo.Ip = strIP
		serviceIp = append(serviceIp, srvinfo)
	}

	recursionInfo, err := d.GetRecursion(ctx, req)
	if err != nil {
		return nil, err
	}

	sysCf := new(pb.SysConfs)
	sysCf.Sys = conf
	sysCf.Ip = serviceIp
	sysCf.Info = recursionInfo
	return sysCf, nil
}

func (d *DnsService) QueryServiceIP(ctx context.Context, in *pb.ReqStatus) (*pb.RespServiceIP, error) {
	logger.Debug("Enter QueryServiceIP.")
	resp := &pb.RespServiceIP{}
	kv := clientv3.NewKV(d.Db)
	//service ip
	tmpKey := "/project_nm/version1_0_0/DDI/domain_manager/config/service_ip"
	gresp, err := kv.Get(context.TODO(), tmpKey, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return resp, err
	}

	var serviceIp []*pb.ServiceIPInfo
	var strKey string
	for _, ev := range gresp.Kvs {
		strKey = string(ev.Key)
		strIP := strKey[len(tmpKey)+1:]
		srvinfo := new(pb.ServiceIPInfo)
		srvinfo.Ip = strIP
		serviceIp = append(serviceIp, srvinfo)
	}
	resp.Ip = serviceIp

	logger.Debug("exit.")
	return resp, nil
}
