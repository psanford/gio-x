// SPDX-License-Identifier: Unlicense OR MIT

package main

// A simple Gio program. See https://gioui.org for more information.

import (
	"image/color"
	"log"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
	"git.sr.ht/~whereswaldon/niotify"

	"gioui.org/font/gofont"
)

func main() {
	go func() {
		w := app.NewWindow()
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	first := true
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			l := material.H1(th, "Hello, Gio")
			maroon := color.RGBA{127, 0, 0, 255}
			l.Color = maroon
			l.Alignment = text.Middle
			l.Layout(gtx)
			e.Frame(gtx.Ops)
			if first {
				first = false
				go func() {
					mgr, err := niotify.NewManager()
					if err != nil {
						log.Printf("manager creation failed: %v", err)
					}
					notif, err := mgr.CreateNotification("hello!", "IS GIO OUT THERE?")
					if err != nil {
						log.Printf("notification send failed: %v", err)
						return
					}
					time.Sleep(time.Second * 10)
					if err := notif.Cancel(); err != nil {
						log.Printf("failed cancelling: %v", err)
					}
				}()
			}
		}
	}
}
