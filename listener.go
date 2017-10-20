package kcpwrapper

import (
	"net"
	"time"

	"github.com/getlantern/cmux"
	"github.com/xtaci/kcp-go"
)

// ListenerConfig configures listening
type ListenerConfig struct {
	CommonConfig
	Listen string `json:"listen"`
}

// Listen listens with the given ListenerConfig
func Listen(cfg *ListenerConfig) (net.Listener, error) {
	cfg.applyDefaults()

	l, err := kcp.ListenWithOptions(cfg.Listen, cfg.block, cfg.DataShard, cfg.ParityShard)
	if err != nil {
		return nil, err
	}
	if err := l.SetDSCP(cfg.DSCP); err != nil {
		log.Errorf("Error setting DSCP: %v", err)
	}
	if err := l.SetReadBuffer(cfg.SockBuf); err != nil {
		log.Errorf("Error setting ReadBuffer: %v", err)
	}
	if err := l.SetWriteBuffer(cfg.SockBuf); err != nil {
		log.Errorf("Error setting WriteBuffer: %v", err)
	}

	return cmux.Listen(&cmux.ListenOpts{
		Listener:          &wrapperListener{Listener: l, cfg: cfg},
		BufferSize:        cfg.SockBuf,
		KeepAliveInterval: time.Duration(cfg.KeepAlive) * time.Second,
	}), nil
}

type wrapperListener struct {
	*kcp.Listener
	cfg *ListenerConfig
}

func (l *wrapperListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.AcceptKCP()
	if err != nil {
		return conn, err
	}

	conn.SetStreamMode(true)
	conn.SetWriteDelay(true)
	conn.SetNoDelay(l.cfg.NoDelay, l.cfg.Interval, l.cfg.Resend, l.cfg.NoCongestion)
	conn.SetMtu(l.cfg.MTU)
	conn.SetWindowSize(l.cfg.SndWnd, l.cfg.RcvWnd)
	conn.SetACKNoDelay(l.cfg.AckNodelay)

	if l.cfg.NoComp {
		return conn, nil
	}

	return wrapSnappy(conn), nil
}
