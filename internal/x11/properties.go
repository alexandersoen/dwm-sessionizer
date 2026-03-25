package x11

import (
	"bytes"

	"github.com/alexandersoen/dwm-sessionizer/internal/model"
	"github.com/alexandersoen/dwm-sessionizer/internal/x11atoms"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type InstanceClassInfo struct {
	Instance string
	Class    string
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

func GetInstanceClass(conn *xgb.Conn, win xproto.Window, wAtoms x11atoms.WindowAtoms) InstanceClassInfo {
	classBytes, err := getProperty(conn, win, wAtoms.WMClass)
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

func GetTitle(conn *xgb.Conn, win xproto.Window, watoms x11atoms.WindowAtoms) string {
	title, titleErr := getTextProperty(conn, win, watoms.NetWMName)
	if title == "" || titleErr != nil {
		title, titleErr = getTextProperty(conn, win, watoms.WMName)
		if titleErr != nil {
			return ""
		}
	}
	return title
}

func GetTransientFor(conn *xgb.Conn, win xproto.Window, watoms x11atoms.WindowAtoms) *xproto.Window {
	transientForRaw, transientErr := getProperty(conn, win, watoms.WMTransientFor)
	if transientErr != nil || len(transientForRaw) < 4 {
		return nil
	}

	parent := xproto.Window(xgb.Get32(transientForRaw))
	return &parent
}

func GetWindowAttributesInfo(conn *xgb.Conn, win xproto.Window) model.WindowAttributesInfo {
	winAttrsCookie := xproto.GetWindowAttributes(conn, win)
	winAttrs, attrErr := winAttrsCookie.Reply()
	if attrErr != nil || winAttrs == nil {
		return model.WindowAttributesInfo{}
	}
	return model.WindowAttributesInfo{
		OverrideRedirect: winAttrs.OverrideRedirect,
		InputOnly:        winAttrs.Class == xproto.WindowClassInputOnly,
	}
}

func GetWindowTypeAtomsForWindow(conn *xgb.Conn, win xproto.Window, watoms x11atoms.WindowAtoms) []xproto.Atom {
	winTypeBytes, typeErr := getProperty(conn, win, watoms.NetWMWindowType)
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
