// the contents of this file is taken from github.com/xtaci/kcptun

// Package kcpwrapper provides a wrapper around kcp that allows the use of
// regular net interfaces like dialing and listening. A lot of the logic is
// based on logic in github.com/xtaci/kcptun.
package kcpwrapper

import (
	"crypto/sha1"
	"math/rand"
	"time"

	"github.com/getlantern/golog"
	kcp "github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// SALT is use for pbkdf2 key expansion
	SALT = "kcp-go"
)

var (
	log = golog.LoggerFor("kcpwrapper")
)

// CommonConfig contains common configuration parameters across dialers and
// listeners.
type CommonConfig struct {
	Key          string `json:"key"`
	Crypt        string `json:"crypt"`
	Mode         string `json:"mode"`
	MTU          int    `json:"mtu"`
	SndWnd       int    `json:"sndwnd"`
	RcvWnd       int    `json:"rcvwnd"`
	DataShard    int    `json:"datashard"`
	ParityShard  int    `json:"parityshard"`
	DSCP         int    `json:"dscp"`
	NoComp       bool   `json:"nocomp"`
	AckNodelay   bool   `json:"acknodelay"`
	NoDelay      int    `json:"nodelay"`
	Interval     int    `json:"interval"`
	Resend       int    `json:"resend"`
	NoCongestion int    `json:"nc"`
	SockBuf      int    `json:"sockbuf"`
	KeepAlive    int    `json:"keepalive"`
	block        kcp.BlockCrypt
}

func (config *CommonConfig) applyDefaults() {
	rand.Seed(int64(time.Now().Nanosecond()))
	switch config.Mode {
	case "normal":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 0, 40, 2, 1
	case "fast":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 0, 30, 2, 1
	case "fast2":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 1, 20, 2, 1
	case "fast3":
		config.NoDelay, config.Interval, config.Resend, config.NoCongestion = 1, 10, 2, 1
	}

	pass := pbkdf2.Key([]byte(config.Key), []byte(SALT), 4096, 32, sha1.New)
	switch config.Crypt {
	case "sm4":
		config.block, _ = kcp.NewSM4BlockCrypt(pass[:16])
	case "tea":
		config.block, _ = kcp.NewTEABlockCrypt(pass[:16])
	case "xor":
		config.block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	case "none":
		config.block, _ = kcp.NewNoneBlockCrypt(pass)
	case "aes-128":
		config.block, _ = kcp.NewAESBlockCrypt(pass[:16])
	case "aes-192":
		config.block, _ = kcp.NewAESBlockCrypt(pass[:24])
	case "blowfish":
		config.block, _ = kcp.NewBlowfishBlockCrypt(pass)
	case "twofish":
		config.block, _ = kcp.NewTwofishBlockCrypt(pass)
	case "cast5":
		config.block, _ = kcp.NewCast5BlockCrypt(pass[:16])
	case "3des":
		config.block, _ = kcp.NewTripleDESBlockCrypt(pass[:24])
	case "xtea":
		config.block, _ = kcp.NewXTEABlockCrypt(pass[:16])
	case "salsa20":
		config.block, _ = kcp.NewSalsa20BlockCrypt(pass)
	default:
		config.Crypt = "aes"
		config.block, _ = kcp.NewAESBlockCrypt(pass)
	}
}
