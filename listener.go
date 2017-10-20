package kcpwrapper

import (
	"net"
	"time"

	"github.com/getlantern/cmux"
	"github.com/getlantern/golog"
	"github.com/xtaci/kcp-go"
)

var (
	log = golog.LoggerFor("kcplistener")
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
		Listener:          l,
		BufferSize:        cfg.SockBuf,
		KeepAliveInterval: time.Duration(cfg.KeepAlive) * time.Second,
	}), nil
}
