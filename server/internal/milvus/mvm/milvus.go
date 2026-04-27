package mvm

import (
	"context"
	"fmt"
	"mayfly-go/pkg/errorx"
	"net"
	"time"

	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"google.golang.org/grpc"
)

// MilvusInfo Milvus 连接信息
type MilvusInfo struct {
	Id                 uint64 `json:"id"`
	Code               string `json:"code"`
	Name               string `json:"name"`
	Host               string `json:"host"`
	ApiKey             string `json:"apiKey"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	Database           string `json:"database"`
	SshTunnelMachineId int    `json:"sshTunnelMachineId"`
}

func (mi *MilvusInfo) Conn() (*MilvusConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cc := &milvusclient.ClientConfig{
		Address: mi.Host,
	}

	if mi.Username != "" && mi.Password != "" {
		cc.Username = mi.Username
		cc.Password = mi.Password
	}

	if mi.ApiKey != "" {
		cc.APIKey = mi.ApiKey
	}

	if mi.Database != "" {
		cc.DBName = mi.Database
	}

	// SSH 隧道
	var opts []grpc.DialOption
	if mi.SshTunnelMachineId > 0 {
		dialer := &MilvusSshDialer{machineId: mi.SshTunnelMachineId}
		opts = append(opts, grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, address)
		}))
	}

	cc.DialOptions = opts

	cli, err := milvusclient.New(ctx, cc)

	if err != nil {
		return nil, errorx.NewBiz(fmt.Sprintf("创建 Milvus 客户端失败：%s", err.Error()))
	}

	// 测试连接并获取版本信息
	service := cli.GetService()
	r, err := service.GetVersion(ctx, &milvuspb.GetVersionRequest{})
	if err != nil {
		_ = cli.Close(ctx)
		return nil, fmt.Errorf("milvus 连接失败：%s", err)
	}
	err = handleRespStatus(r.GetStatus(), err)
	if err != nil {
		_ = cli.Close(ctx)
		return nil, fmt.Errorf("milvus 连接失败：%s", err)
	}

	return &MilvusConn{
		Id:      mi.Id,
		cli:     cli,
		info:    mi,
		Service: service,
		DbName:  mi.Database,
		Version: r.Version,
	}, nil

}

func getConnId(id uint64, database string) string {
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("milvus:%d:%s", id, database)
}
