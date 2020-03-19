package dhcp

import (
	"context"

	"reyzar.com/server-api/pkg/logger"
	"reyzar.com/server-api/pkg/rpc/dhcpclient"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

func (ds *DhcpService) GetDeviceInfo(ctx context.Context, in *pb.GetDeviceInfoReq) (*pb.GetDeviceInfoRsp, error) {
	logger.Debug("Enter GetDeviceInfo.")
	resp := &pb.GetDeviceInfoRsp{}

	req := &dhcpclient.GetDeviceInfoReq{
		IpAddr: in.IpAddr,
	}

	for i := 0; i < len(ds.RdhcpAgentConn); i++ {
		c := dhcpclient.NewDevicePerformanceClient(ds.RdhcpAgentConn[i])
		rsp, err := c.GetDeviceInfo(ctx, req)
		if err != nil {
			logger.Error(err)
			continue
		}
		if rsp.IpAddr != "" {
			tmpStat := new(pb.DeviceStatInfo)
			tmpStat.IpAddr = rsp.IpAddr
			tmpStat.CpuUsageRate = rsp.CpuUsageRate
			tmpStat.MemoryUsageRate = rsp.MemoryUsageRate
			tmpStat.DiskUsageRate = rsp.DiskUsageRate
			resp.StatInfo = append(resp.StatInfo, tmpStat)
		}
	}

	logger.Debug("Exit.")
	return resp, nil
}
