package x11

import (
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func Connect() *xgb.Conn {
	conn, err := xgb.NewConn()
	if err != nil {
		log.Fatalf("connect to X11: %v", err)
	}

	return conn
}

func DefaultScreen(conn *xgb.Conn) xproto.ScreenInfo {
	setup := xproto.Setup(conn)
	screen := *setup.DefaultScreen(conn)
	return screen
}
