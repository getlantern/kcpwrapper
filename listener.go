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

	var listener net.Listener = l
	if !cfg.NoComp {
		listener = &snappyListener{l}
	}

	return cmux.Listen(&cmux.ListenOpts{
		Listener:          listener,
		BufferSize:        cfg.SockBuf,
		KeepAliveInterval: time.Duration(cfg.KeepAlive) * time.Second,
	}), nil
}

type snappyListener struct {
	net.Listener
}

func (l *snappyListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return conn, err
	}
	return wrapSnappy(conn), err
}
