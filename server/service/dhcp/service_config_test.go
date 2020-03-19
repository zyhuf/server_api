package dhcp

import (
	"context"
	"fmt"
	"log"
	"testing"

	"google.golang.org/grpc"
	pb "reyzar.com/server-api/pkg/rpc/dhcpserver"
)

const address = "127.0.0.1:50057"

func ConstructV4Sub() *pb.V4SubInfo {
	V4Sub := new(pb.V4SubInfo)

	var segmentInfo []*pb.NetworkSegmentInfo
	stInfo := new(pb.NetworkSegmentInfo)
	stInfo.Index = 1
	stInfo.Net = "192.168.0.0/16"
	segmentInfo = append(segmentInfo, stInfo)

	stInfo = new(pb.NetworkSegmentInfo)
	stInfo.Index = 2
	stInfo.Net = "172.16.0.0/16"
	segmentInfo = append(segmentInfo, stInfo)

	V4Sub.NetSegment = append(V4Sub.NetSegment, segmentInfo...)

	var poolInfo []*pb.PoolInfo
	tmpPool := new(pb.PoolInfo)
	tmpPool.Index = 1
	tmpPool.Beg = "192.168.0.1"
	tmpPool.End = "192.168.0.10"
	poolInfo = append(poolInfo, tmpPool)
	tmpPool = new(pb.PoolInfo)
	tmpPool.Index = 2
	tmpPool.Beg = "172.16.0.1"
	tmpPool.End = "172.16.0.10"
	poolInfo = append(poolInfo, tmpPool)
	V4Sub.Pool = append(V4Sub.Pool, poolInfo...)

	return V4Sub
}

func ConstructV6Sub() *pb.V6SubInfo {
	V6Sub := new(pb.V6SubInfo)

	var segmentInfo []*pb.NetworkSegmentInfo
	stInfo := new(pb.NetworkSegmentInfo)
	stInfo.Index = 1
	stInfo.Net = "240c::/16"
	segmentInfo = append(segmentInfo, stInfo)

	stInfo = new(pb.NetworkSegmentInfo)
	stInfo.Index = 2
	stInfo.Net = "fe80::/16"
	segmentInfo = append(segmentInfo, stInfo)

	V6Sub.NetSegment = append(V6Sub.NetSegment, segmentInfo...)

	var poolInfo []*pb.PoolInfo
	tmpPool := new(pb.PoolInfo)
	tmpPool.Index = 1
	tmpPool.Beg = "240c::1"
	tmpPool.End = "240c::100"
	poolInfo = append(poolInfo, tmpPool)
	tmpPool = new(pb.PoolInfo)
	tmpPool.Index = 2
	tmpPool.Beg = "fe80::1"
	tmpPool.End = "fe80::100"
	poolInfo = append(poolInfo, tmpPool)
	V6Sub.Pool = append(V6Sub.Pool, poolInfo...)

	PDPrefix := new(pb.PDPrefixInfo)
	PDPrefix.StartPrefix = "2411:11"
	PDPrefix.EndPrefix = "2411:99"
	V6Sub.PdInfo = PDPrefix

	return V6Sub
}

func ConstructSubnet() []*pb.SubnetInfo {
	var subnetInfo []*pb.SubnetInfo

	info := new(pb.SubnetInfo)
	info.Index = 7001
	info.Name = "中山"
	info.VlanId = 12
	info.Type = pb.SubnetType_access_subnet
	info.SubnetValid = 300
	info.BwType = pb.ListType_black_list
	info.V4Sub = ConstructV4Sub()
	info.V6Sub = ConstructV6Sub()
	subnetInfo = append(subnetInfo, info)

	info = new(pb.SubnetInfo)
	info.Index = 7002
	info.Name = "暨南"
	info.VlanId = 13
	info.Type = pb.SubnetType_auth_subnet
	info.BwType = pb.ListType_white_list
	info.V4Sub = ConstructV4Sub()
	info.V6Sub = ConstructV6Sub()
	subnetInfo = append(subnetInfo, info)

	return subnetInfo
}

func DelSubnet() []*pb.SubnetInfo {
	var subnetInfo []*pb.SubnetInfo

	info := new(pb.SubnetInfo)
	info.Index = 7001
	subnetInfo = append(subnetInfo, info)

	info = new(pb.SubnetInfo)
	info.Index = 7002
	subnetInfo = append(subnetInfo, info)

	return subnetInfo
}

func TestOperateSubnet(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewServiceConfigClient(conn)

	subnetInfo := ConstructSubnet()
	//subnetInfo := DelSubnet()

	req := &pb.SubnetInfos{
		OperateType: pb.OperationType_add_type,
		Ev:          subnetInfo,
	}

	reply, err := c.OperateSubnet(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(reply)
}

func TestGetSubnet(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewServiceConfigClient(conn)
	req := &pb.ReqStatus{
		Index: 7001,
	}

	reply, err := c.GetSubnet(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(reply)
}

func ConstructStatic() []*pb.IpStaticInfo {
	var staticInfo []*pb.IpStaticInfo

	info := new(pb.IpStaticInfo)
	info.Index = 8001
	info.SubnetId = 1213
	info.Mac = "00:0c:29:5e:1e:71"
	info.Duid = "12321321312"
	info.V4Ip = "10.2.21.63"
	info.V6Ip = "240c::11"
	staticInfo = append(staticInfo, info)

	info = new(pb.IpStaticInfo)
	info.Index = 8002
	info.SubnetId = 1211
	info.Mac = "00:0c:29:5e:1e:d1"
	info.Duid = "12321321312"
	info.V4Ip = "10.2.21.39"
	info.V6Ip = "240c::12"
	staticInfo = append(staticInfo, info)

	return staticInfo
}

func DelStatic() []*pb.IpStaticInfo {
	var staticInfo []*pb.IpStaticInfo

	info := new(pb.IpStaticInfo)
	info.Index = 8001
	staticInfo = append(staticInfo, info)

	info = new(pb.IpStaticInfo)
	info.Index = 8002
	staticInfo = append(staticInfo, info)

	return staticInfo
}

func TestOperateStatic(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewServiceConfigClient(conn)

	static := ConstructStatic()
	req := &pb.IpStaticInfos{
		OperateType: pb.OperationType_add_type,
		Ev:          static,
	}

	reply, err := c.OperateStatic(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(reply)
}

func TestGetStatic(t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	c := pb.NewServiceConfigClient(conn)
	req := &pb.ReqStatus{
		Index: 8001,
	}

	reply, err := c.GetStatic(context.TODO(), req)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(reply)
}
