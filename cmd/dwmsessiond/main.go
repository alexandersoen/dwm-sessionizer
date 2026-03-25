package main

import (
	"fmt"

	"github.com/alexandersoen/dwm-sessionizer/internal/x11"
	"github.com/alexandersoen/dwm-sessionizer/internal/x11atoms"
)

func main() {
	conn := x11.Connect()
	defer conn.Close()

	screen := x11.DefaultScreen(conn)
	root := screen.Root

	watoms := x11atoms.InternAtoms(conn)

	fmt.Printf("root window: %d\n", root)
	fmt.Printf("size: %dx%d\n", screen.WidthInPixels, screen.HeightInPixels)
	fmt.Println()

	windows := x11.ScanRootWindows(conn, root)
	infos := x11.InspectWindows(conn, windows, watoms)

	for _, info := range infos {
		fmt.Println(info)
		fmt.Println()
	}
}
