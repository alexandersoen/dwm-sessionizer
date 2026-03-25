package main

import (
	"fmt"
	"log"

	"github.com/alexandersoen/dwm-sessionizer/internal/model"
	"github.com/alexandersoen/dwm-sessionizer/internal/x11"
	"github.com/alexandersoen/dwm-sessionizer/internal/x11atoms"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func containsIgnoredWindowType(typeAtoms []xproto.Atom, winTypeAtoms x11atoms.WindowTypeAtoms) bool {
	for _, typeAtom := range typeAtoms {
		if typeAtom == winTypeAtoms.Dock ||
			typeAtom == winTypeAtoms.Notification ||
			typeAtom == winTypeAtoms.Tooltip ||
			typeAtom == winTypeAtoms.Popup ||
			typeAtom == winTypeAtoms.Desktop ||
			typeAtom == winTypeAtoms.Dropdown {
			return true
		}
	}
	return false
}

func classifyWindow(winInfo model.WindowInfo, winTypeAtoms x11atoms.WindowTypeAtoms) model.WindowClassification {
	if winInfo.OverrideRedirect || winInfo.InputOnly {
		return model.Ignored
	}
	if containsIgnoredWindowType(winInfo.WindowTypeAtoms, winTypeAtoms) {
		return model.Ignored
	}
	if winInfo.TransientFor != nil {
		return model.Transient
	}
	return model.Managed
}

func getWindowInfo(conn *xgb.Conn, win xproto.Window, watoms x11atoms.WindowAtoms) model.WindowInfo {
	attrInfo := x11.GetWindowAttributesInfo(conn, win)
	winInfo := model.WindowInfo{
		ID:               win,
		Title:            x11.GetTitle(conn, win, watoms),
		TransientFor:     x11.GetTransientFor(conn, win, watoms),
		OverrideRedirect: attrInfo.OverrideRedirect,
		InputOnly:        attrInfo.InputOnly,
		WindowTypeAtoms:  x11.GetWindowTypeAtomsForWindow(conn, win, watoms),
	}
	instanceClass := x11.GetInstanceClass(conn, win, watoms)
	winInfo.Instance = instanceClass.Instance
	winInfo.Class = instanceClass.Class
	winInfo.Classification = classifyWindow(winInfo, watoms.WindowTypes)
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

	watoms := x11atoms.InternAtoms(conn)

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
		fmt.Println(winInfo)
		fmt.Println()
	}
}
