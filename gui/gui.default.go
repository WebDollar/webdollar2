//go:build !wasm
// +build !wasm

package gui

import (
	"pandora-pay/gui/gui_interactive"
)

func create_gui() (err error) {
	if GUI, err = gui_interactive.CreateGUIInteractive(); err != nil {
		return
	}
	return
}
