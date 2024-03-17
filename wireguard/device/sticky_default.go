//go:build !linux

package device

import (
	"github.com/hossinasaadi/warp-plus/wireguard/conn"
	"github.com/hossinasaadi/warp-plus/wireguard/rwcancel"
)

func (device *Device) startRouteListener(bind conn.Bind) (*rwcancel.RWCancel, error) {
	return nil, nil
}
