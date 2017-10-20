package kcpwrapper

import (
	"context"
	"net"
	"time"

	"github.com/getlantern/cmux"
	"github.com/getlantern/errors"
	"github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

// DialerConfig configures dialing
type DialerConfig struct {
	CommonConfig
	Conn        int `json:"conn"`
	AutoExpire  int `json:"autoexpire"`
	ScavengeTTL int `json:"scavengettl"`
}

// Dialer creates a new dialer function
func Dialer(cfg *DialerConfig) func(ctx context.Context, network, addr string) (net.Conn, error) {
	cfg.applyDefaults()

	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = cfg.SockBuf
	smuxConfig.KeepAliveInterval = time.Duration(cfg.KeepAlive) * time.Second

	dialKCP := func(ctx context.Context, network, addr string) (net.Conn, error) {
		kcpconn, err := kcp.DialWithOptions(addr, cfg.block, cfg.DataShard, cfg.ParityShard)
		if err != nil {
			return nil, errors.New("Unable to dial KCP: %v", err)
		}

		kcpconn.SetStreamMode(true)
		kcpconn.SetWriteDelay(true)
		kcpconn.SetNoDelay(cfg.NoDelay, cfg.Interval, cfg.Resend, cfg.NoCongestion)
		kcpconn.SetWindowSize(cfg.SndWnd, cfg.RcvWnd)
		kcpconn.SetMtu(cfg.MTU)
		kcpconn.SetACKNoDelay(cfg.AckNodelay)

		if err := kcpconn.SetDSCP(cfg.DSCP); err != nil {
			log.Errorf("Unable to set DSCP: %v", err)
		}
		if err := kcpconn.SetReadBuffer(cfg.SockBuf); err != nil {
			log.Errorf("Unable to set ReadBuffer: %v", err)
		}
		if err := kcpconn.SetWriteBuffer(cfg.SockBuf); err != nil {
			log.Errorf("Unable to set WriteBuffer:: %v", err)
		}

		return kcpconn, nil
	}

	return cmux.Dialer(&cmux.DialerOpts{
		Dial:              dialKCP,
		PoolSize:          cfg.Conn,
		BufferSize:        cfg.SockBuf,
		KeepAliveInterval: time.Duration(cfg.KeepAlive) * time.Second,
	})
}
