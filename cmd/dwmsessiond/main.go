package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type WindowTypeAtoms struct {
	dock         xproto.Atom
	desktop      xproto.Atom
	notification xproto.Atom
	tooltip      xproto.Atom
	popup        xproto.Atom
	dropdown     xproto.Atom
	dialog       xproto.Atom
	normal       xproto.Atom
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
		dock:         dock,
		desktop:      desktop,
		notification: notification,
		tooltip:      tooltip,
		popup:        popup,
		dropdown:     dropdown,
		dialog:       dialog,
		normal:       normal,
	}
}

type WindowAtoms struct {
	netWmName       xproto.Atom
	wmName          xproto.Atom
	wmClass         xproto.Atom
	wmTransientFor  xproto.Atom
	netWmWindowType xproto.Atom
	winTypeAtoms    WindowTypeAtoms
}

func internAtom(conn *xgb.Conn, name string) xproto.Atom {
	atomCookie := xproto.InternAtom(conn, false, uint16(len(name)), name)
	atomReply, err := atomCookie.Reply()
	if err != nil {
		log.Fatalf("intern atom: %v", err)
	}

	return atomReply.Atom
}

func internAtoms(conn *xgb.Conn) WindowAtoms {
	netWmName := internAtom(conn, "_NET_WM_NAME")
	wmName := internAtom(conn, "WM_NAME")
	wmClass := internAtom(conn, "WM_CLASS")
	wmTransientFor := internAtom(conn, "WM_TRANSIENT_FOR")
	netWmWindowType := internAtom(conn, "_NET_WM_WINDOW_TYPE")
	winTypeAtoms := internWindowType(conn)

	return WindowAtoms{
		netWmName:       netWmName,
		wmName:          wmName,
		wmClass:         wmClass,
		wmTransientFor:  wmTransientFor,
		netWmWindowType: netWmWindowType,
		winTypeAtoms:    winTypeAtoms,
	}

}

type WindowClassification int

const (
	Ignored WindowClassification = iota + 1
	Managed
	Transient
)

func (wt WindowClassification) String() string {
	switch wt {
	case Ignored:
		return "Ignored"
	case Managed:
		return "Managed"
	case Transient:
		return "Transient"
	}
	return ""
}

type WindowInfo struct {
	ID               xproto.Window
	Title            string
	Instance         string
	Class            string
	TransientFor     *xproto.Window
	OverrideRedirect bool
	InputOnly        bool
	WindowTypeAtoms  []xproto.Atom
	Classification   WindowClassification
}

type WindowAttributesInfo struct {
	OverrideRedirect bool
	InputOnly        bool
}

type InstanceClassInfo struct {
	Instance string
	Class    string
}

func printWindowInfo(winInfo WindowInfo) {
	fmt.Printf("ID: %d\n", winInfo.ID)
	fmt.Printf("Title: %s\n", winInfo.Title)
	fmt.Printf("Instance: %s\n", winInfo.Instance)
	fmt.Printf("Class: %s\n", winInfo.Class)
	if winInfo.TransientFor == nil {
		fmt.Printf("TransientFor: none\n")
	} else {
		fmt.Printf("TransientFor: %d\n", *winInfo.TransientFor)
	}
	fmt.Printf("OverrideRedirect: %t\n", winInfo.OverrideRedirect)
	fmt.Printf("InputOnly: %t\n", winInfo.InputOnly)
	fmt.Printf("WindowTypeAtoms: %v\n", winInfo.WindowTypeAtoms)
	fmt.Printf("Classification: %s\n", winInfo.Classification)
}

func getProperty(conn *xgb.Conn, win xproto.Window, prop xproto.Atom) ([]byte, error) {
	cookie := xproto.GetProperty(conn, false, win, prop, xproto.AtomAny, 0, 1024)
	reply, err := cookie.Reply()
	if err != nil {
		return nil, err
	}
	return reply.Value, nil
}

func getTextProperty(conn *xgb.Conn, win xproto.Window, prop xproto.Atom) (string, error) {
	value, err := getProperty(conn, win, prop)
	if err != nil {
		return "", err
	}
	if len(value) == 0 {
		return "", nil
	}
	return string(value), nil
}

func getInstanceClass(conn *xgb.Conn, win xproto.Window, wAtoms WindowAtoms) InstanceClassInfo {
	classBytes, err := getProperty(conn, win, wAtoms.wmClass)
	if err != nil || len(classBytes) == 0 {
		return InstanceClassInfo{}
	}

	// Drop trailing empties
	parts := bytes.Split(classBytes, []byte{0})
	for len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}

	switch len(parts) {
	case 0:
		return InstanceClassInfo{}
	case 1:
		return InstanceClassInfo{Instance: string(parts[0])}
	default:
		return InstanceClassInfo{
			Instance: string(parts[0]),
			Class:    string(parts[1]),
		}
	}
}

func getTitle(conn *xgb.Conn, win xproto.Window, watoms WindowAtoms) string {
	title, titleErr := getTextProperty(conn, win, watoms.netWmName)
	if title == "" || titleErr != nil {
		title, titleErr = getTextProperty(conn, win, watoms.wmName)
		if titleErr != nil {
			return ""
		}
	}
	return title
}

func getTransientFor(conn *xgb.Conn, win xproto.Window, watoms WindowAtoms) *xproto.Window {
	transientForRaw, transientErr := getProperty(conn, win, watoms.wmTransientFor)
	if transientErr != nil || len(transientForRaw) < 4 {
		return nil
	}

	parent := xproto.Window(xgb.Get32(transientForRaw))
	return &parent
}

func getWindowAttributesInfo(conn *xgb.Conn, win xproto.Window) WindowAttributesInfo {
	winAttrsCookie := xproto.GetWindowAttributes(conn, win)
	winAttrs, attrErr := winAttrsCookie.Reply()
	if attrErr != nil || winAttrs == nil {
		return WindowAttributesInfo{}
	}
	return WindowAttributesInfo{
		OverrideRedirect: winAttrs.OverrideRedirect,
		InputOnly:        winAttrs.Class == xproto.WindowClassInputOnly,
	}
}

func getWindowTypeAtomsForWindow(conn *xgb.Conn, win xproto.Window, watoms WindowAtoms) []xproto.Atom {
	winTypeBytes, typeErr := getProperty(conn, win, watoms.netWmWindowType)
	if typeErr != nil || len(winTypeBytes) == 0 || len(winTypeBytes)%4 != 0 {
		return nil
	}

	var typeAtoms []xproto.Atom
	for i := 0; i+4 <= len(winTypeBytes); i += 4 {
		typeAtom := xproto.Atom(xgb.Get32(winTypeBytes[i : i+4]))
		typeAtoms = append(typeAtoms, typeAtom)
	}
	return typeAtoms
}

func containsIgnoredWindowType(typeAtoms []xproto.Atom, winTypeAtoms WindowTypeAtoms) bool {
	for _, typeAtom := range typeAtoms {
		if typeAtom == winTypeAtoms.dock ||
			typeAtom == winTypeAtoms.notification ||
			typeAtom == winTypeAtoms.tooltip ||
			typeAtom == winTypeAtoms.popup ||
			typeAtom == winTypeAtoms.desktop ||
			typeAtom == winTypeAtoms.dropdown {
			return true
		}
	}
	return false
}

func classifyWindow(winInfo WindowInfo, winTypeAtoms WindowTypeAtoms) WindowClassification {
	if winInfo.OverrideRedirect || winInfo.InputOnly {
		return Ignored
	}
	if containsIgnoredWindowType(winInfo.WindowTypeAtoms, winTypeAtoms) {
		return Ignored
	}
	if winInfo.TransientFor != nil {
		return Transient
	}
	return Managed
}

func getWindowInfo(conn *xgb.Conn, win xproto.Window, watoms WindowAtoms) WindowInfo {
	attrInfo := getWindowAttributesInfo(conn, win)
	winInfo := WindowInfo{
		ID:               win,
		Title:            getTitle(conn, win, watoms),
		TransientFor:     getTransientFor(conn, win, watoms),
		OverrideRedirect: attrInfo.OverrideRedirect,
		InputOnly:        attrInfo.InputOnly,
		WindowTypeAtoms:  getWindowTypeAtomsForWindow(conn, win, watoms),
	}
	instanceClass := getInstanceClass(conn, win, watoms)
	winInfo.Instance = instanceClass.Instance
	winInfo.Class = instanceClass.Class
	winInfo.Classification = classifyWindow(winInfo, watoms.winTypeAtoms)
	return winInfo
}

func main() {
	conn, err := xgb.NewConn()
	if err != nil {
		log.Fatalf("connect to X11: %v", err)
	}
	defer conn.Close()

	setup := xproto.Setup(conn)
	screen := setup.DefaultScreen(conn)

	watoms := internAtoms(conn)

	fmt.Printf("root window: %d\n", screen.Root)
	fmt.Printf("size: %dx%d\n", screen.WidthInPixels, screen.HeightInPixels)
	fmt.Println()

	cookie := xproto.QueryTree(conn, screen.Root)
	reply, err := cookie.Reply()
	if err != nil {
		log.Fatalf("query tree: %v", err)
	}

	for _, win := range reply.Children {
		winInfo := getWindowInfo(conn, win, watoms)
		printWindowInfo(winInfo)
		fmt.Println()
	}
}
