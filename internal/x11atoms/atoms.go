package x11atoms

import (
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type WindowTypeAtoms struct {
	Dock         xproto.Atom
	Desktop      xproto.Atom
	Notification xproto.Atom
	Tooltip      xproto.Atom
	Popup        xproto.Atom
	Dropdown     xproto.Atom
	Dialog       xproto.Atom
	Normal       xproto.Atom
}

type WindowAtoms struct {
	NetWMName       xproto.Atom
	WMName          xproto.Atom
	WMClass         xproto.Atom
	WMTransientFor  xproto.Atom
	NetWMWindowType xproto.Atom
	WindowTypes     WindowTypeAtoms
}

func internAtom(conn *xgb.Conn, name string) xproto.Atom {
	atomCookie := xproto.InternAtom(conn, false, uint16(len(name)), name)
	atomReply, err := atomCookie.Reply()
	if err != nil {
		log.Fatalf("intern atom: %v", err)
	}

	return atomReply.Atom
}

func internWindowType(conn *xgb.Conn) WindowTypeAtoms {
	dock := internAtom(conn, "_NET_WM_WINDOW_TYPE_DOCK")
	desktop := internAtom(conn, "_NET_WM_WINDOW_TYPE_DESKTOP")
	notification := internAtom(conn, "_NET_WM_WINDOW_TYPE_NOTIFICATION")
	tooltip := internAtom(conn, "_NET_WM_WINDOW_TYPE_TOOLTIP")
	popup := internAtom(conn, "_NET_WM_WINDOW_TYPE_POPUP_MENU")
	dropdown := internAtom(conn, "_NET_WM_WINDOW_TYPE_DROPDOWN_MENU")
	dialog := internAtom(conn, "_NET_WM_WINDOW_TYPE_DIALOG")
	normal := internAtom(conn, "_NET_WM_WINDOW_TYPE_NORMAL")

	return WindowTypeAtoms{
		Dock:         dock,
		Desktop:      desktop,
		Notification: notification,
		Tooltip:      tooltip,
		Popup:        popup,
		Dropdown:     dropdown,
		Dialog:       dialog,
		Normal:       normal,
	}
}

func InternAtoms(conn *xgb.Conn) WindowAtoms {
	netWmName := internAtom(conn, "_NET_WM_NAME")
	wmName := internAtom(conn, "WM_NAME")
	wmClass := internAtom(conn, "WM_CLASS")
	wmTransientFor := internAtom(conn, "WM_TRANSIENT_FOR")
	netWmWindowType := internAtom(conn, "_NET_WM_WINDOW_TYPE")
	winTypeAtoms := internWindowType(conn)

	return WindowAtoms{
		NetWMName:       netWmName,
		WMName:          wmName,
		WMClass:         wmClass,
		WMTransientFor:  wmTransientFor,
		NetWMWindowType: netWmWindowType,
		WindowTypes:     winTypeAtoms,
	}

}
