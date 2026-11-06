package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestDiscardKeybinding(t *testing.T) {
	event := tcell.NewEventKey(tcell.KeyF8, 0, tcell.ModNone)
	if got := Keymaps.Group(HomeGroup).Resolve(event).String(); got != "DiscardChanges" {
		t.Fatalf("F8 command = %q, want DiscardChanges", got)
	}
	for _, bind := range Keymaps.Group(HomeGroup) {
		if bind.Key.Code == tcell.KeyF8 && bind.Description == "" {
			t.Fatal("discard shortcut must be discoverable in help")
		}
	}
	if got := Keymaps.Group(EditorGroup).Resolve(tcell.NewEventKey(tcell.KeyCtrlR, 0, tcell.ModNone)).String(); got != "Execute" {
		t.Fatalf("editor Ctrl+R command = %q, want Execute", got)
	}
}
