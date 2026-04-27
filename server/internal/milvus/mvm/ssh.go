package mvm

import (
	"context"
	machineapp "mayfly-go/internal/machine/application"
	"mayfly-go/pkg/utils/netx"
	"net"
)

// MilvusSshDialer SSH 隧道 Dialer
type MilvusSshDialer struct {
	machineId int
}

func (sd *MilvusSshDialer) DialContext(ctx context.Context, address string) (net.Conn, error) {
	stm, err := machineapp.GetMachineApp().GetSshTunnelMachine(ctx, sd.machineId)
	if err != nil {
		return nil, err
	}
	if sshConn, err := stm.GetDialConn("tcp", address); err == nil {
		// 使用 WrapSshConn 包装，避免 deadline 不支持的问题
		return &netx.WrapSshConn{Conn: sshConn}, nil
	} else {
		return nil, err
	}
}
