package x11

import (
	"log"

	"github.com/alexandersoen/dwm-sessionizer/internal/model"
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
	attrInfo := GetWindowAttributesInfo(conn, win)
	winInfo := model.WindowInfo{
		ID:               win,
		Title:            GetTitle(conn, win, watoms),
		TransientFor:     GetTransientFor(conn, win, watoms),
		OverrideRedirect: attrInfo.OverrideRedirect,
		InputOnly:        attrInfo.InputOnly,
		WindowTypeAtoms:  GetWindowTypeAtomsForWindow(conn, win, watoms),
	}
	instanceClass := GetInstanceClass(conn, win, watoms)
	winInfo.Instance = instanceClass.Instance
	winInfo.Class = instanceClass.Class
	winInfo.Classification = classifyWindow(winInfo, watoms.WindowTypes)
	return winInfo
}

func ScanRootWindows(conn *xgb.Conn, root xproto.Window) []xproto.Window {

	cookie := xproto.QueryTree(conn, root)
	reply, err := cookie.Reply()
	if err != nil {
		log.Fatalf("query tree: %v", err)
	}

	return reply.Children
}

func InspectWindows(conn *xgb.Conn, windows []xproto.Window, watoms x11atoms.WindowAtoms) []model.WindowInfo {
	infos := make([]model.WindowInfo, len(windows))
	for i, win := range windows {
		infos[i] = getWindowInfo(conn, win, watoms)
	}

	return infos
}
