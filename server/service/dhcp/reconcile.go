package dhcp

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"reyzar.com/server-api/pkg/logger"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

func (ds *DhcpService) SetAddressInspectCfg(ctx context.Context, in *pb.InspectCfgReq) (*pb.RespResult, error) {
	logger.Debug("Enter SetServerCfg.")
	result := &pb.RespResult{}

	key := ds.KeyPrefix + "address_manager/reconcile/inspect_cycle"
	value := fmt.Sprintf("%v", in.InspectCycle)

	kvc := clientv3.NewKV(ds.Db)
	_, err := kvc.Txn(context.TODO()).Then(clientv3.OpPut(key, value)).Commit()
	if err != nil {
		logger.Error(err)
		result.ResultCode = ErrorList[0].ErrCode
		result.Description = ErrorList[0].ErrDesc
	}

	logger.Debug("Exit.")

	return result, nil
}
