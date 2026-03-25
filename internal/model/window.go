package model

import (
	"fmt"
	"strings"

	"github.com/jezek/xgb/xproto"
)

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

func (winInfo WindowInfo) String() string {
	var transientFor string
	if winInfo.TransientFor == nil {
		transientFor = "none"
	} else {
		transientFor = fmt.Sprintf("%d", *winInfo.TransientFor)
	}

	lines := []string{
		fmt.Sprintf("ID: %d", winInfo.ID),
		fmt.Sprintf("Title: %s", winInfo.Title),
		fmt.Sprintf("Instance: %s", winInfo.Instance),
		fmt.Sprintf("Class: %s", winInfo.Class),
		fmt.Sprintf("TransientFor: %s", transientFor),
		fmt.Sprintf("OverrideRedirect: %t", winInfo.OverrideRedirect),
		fmt.Sprintf("InputOnly: %t", winInfo.InputOnly),
		fmt.Sprintf("WindowTypeAtoms: %v", winInfo.WindowTypeAtoms),
		fmt.Sprintf("Classification: %s", winInfo.Classification),
	}

	return strings.Join(lines, "\n")
}
