package components

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func sidebarTestUI(t *testing.T, height int, values ...string) (*Sidebar, tcell.SimulationScreen) {
	t.Helper()
	previous := App.Application
	App.Application = tview.NewApplication()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(100, 40)
	App.SetScreen(screen)
	sidebar := NewSidebar("mysql", false)
	sidebar.SetRect(50, 7, 30, height)
	for i, value := range values {
		sidebar.AddField(fmt.Sprintf("field%d", i), value, 28, i == 1)
	}
	App.SetRoot(sidebar, false)
	if len(values) > 0 {
		App.SetFocus(sidebar)
	}
	sidebar.Draw(screen)
	t.Cleanup(func() {
		screen.Fini()
		App.Application = previous
	})
	return sidebar, screen
}

func sidebarMouse(sidebar *Sidebar, action tview.MouseAction, x, y int) bool {
	consumed, _ := sidebar.MouseHandler()(action, tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
	return consumed
}

func assertSidebarSelection(t *testing.T, sidebar *Sidebar, screen tcell.Screen, want int) {
	t.Helper()
	sidebar.Draw(screen)
	field := sidebar.Flex.GetItem(want).(*tview.TextArea)
	if sidebar.GetCurrentFieldIndex() != want || App.GetFocus() != field {
		t.Fatalf("selected index = %d, focus = %p; want field %d (%p)", sidebar.GetCurrentFieldIndex(), App.GetFocus(), want, field)
	}
	_, top, _, height := sidebar.GetInnerRect()
	_, y, _, h := field.GetRect()
	if y < top || h != min(height, sidebar.FieldParameters[want].Height) || y+h > top+height {
		t.Fatalf("selected field %d is clipped: y=%d height=%d, viewport=%d..%d", want, y, h, top, top+height)
	}
	for i := 0; i < sidebar.Flex.GetItemCount(); i++ {
		_, y, _, h := sidebar.Flex.GetItem(i).GetRect()
		if h > 0 && (y < top || y+h > top+height) {
			t.Fatalf("field %d draws outside sidebar: y=%d height=%d", i, y, h)
		}
	}
}

func TestSidebarWheelNavigatesFields(t *testing.T) {
	sidebar, screen := sidebarTestUI(t, 10, "one", strings.Repeat("wide value ", 30), "three", "four")
	for i := 1; i <= 3; i++ {
		if !sidebarMouse(sidebar, tview.MouseScrollDown, 55, 9) {
			t.Fatal("sidebar did not consume mouse wheel")
		}
		assertSidebarSelection(t, sidebar, screen, i)
	}
	sidebarMouse(sidebar, tview.MouseScrollDown, 55, 9)
	assertSidebarSelection(t, sidebar, screen, 3)
	for i := 2; i >= 0; i-- {
		sidebarMouse(sidebar, tview.MouseScrollUp, 55, 9)
		assertSidebarSelection(t, sidebar, screen, i)
	}
	sidebarMouse(sidebar, tview.MouseScrollUp, 55, 9)
	assertSidebarSelection(t, sidebar, screen, 0)
	if sidebarMouse(sidebar, tview.MouseScrollDown, 10, 9) || sidebar.GetCurrentFieldIndex() != 0 {
		t.Fatal("wheel outside sidebar changed selection")
	}
}

func TestSidebarClickSelectsFieldWithoutEditing(t *testing.T) {
	for _, border := range []bool{false, true} {
		t.Run(fmt.Sprintf("border=%t", border), func(t *testing.T) {
			sidebar, screen := sidebarTestUI(t, 15, "first", "second", "third")
			field := sidebar.Flex.GetItem(1).(*tview.TextArea)
			x, y, _, _ := field.GetRect()
			if !border {
				x++
				y++
			}
			if !sidebarMouse(sidebar, tview.MouseLeftClick, x, y) {
				t.Fatal("click not handled")
			}
			assertSidebarSelection(t, sidebar, screen, 1)
			if !field.GetDisabled() || field.GetText() != "second" {
				t.Fatal("click changed field content or enabled editing")
			}
			sidebar.FocusNextField()
			assertSidebarSelection(t, sidebar, screen, 2)
		})
	}
}

func TestSidebarKeyboardKeepsTallFieldsVisible(t *testing.T) {
	long := strings.Repeat("x", 200)
	sidebar, screen := sidebarTestUI(t, 10, "one", long, long, "four", long)
	for i := 1; i <= 4; i++ {
		sidebar.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
		assertSidebarSelection(t, sidebar, screen, i)
	}
	for i := 3; i >= 0; i-- {
		sidebar.FocusPreviousField()
		assertSidebarSelection(t, sidebar, screen, i)
	}
	sidebar.FocusLastField()
	assertSidebarSelection(t, sidebar, screen, 4)
	sidebar.FocusFirstField()
	assertSidebarSelection(t, sidebar, screen, 0)
	sidebar.FocusField(3)
	assertSidebarSelection(t, sidebar, screen, 3)
	sidebar.SetRect(50, 12, 30, 5)
	assertSidebarSelection(t, sidebar, screen, 3)
}

func TestSidebarClickDoesNotRetargetAfterRedraw(t *testing.T) {
	sidebar, screen := sidebarTestUI(t, 10, "one", "two", "three", "four")
	field := sidebar.Flex.GetItem(2).(*tview.TextArea)
	x, y, _, _ := field.GetRect()
	root := tview.NewPages().AddPage("sidebar", sidebar, false, true)
	root.SetRect(0, 0, 100, 40)
	root.Draw(screen)
	setFocus := func(p tview.Primitive) { App.SetFocus(p) }
	mouse := func(action tview.MouseAction) {
		root.MouseHandler()(action, tcell.NewEventMouse(x+1, y+1, tcell.ButtonNone, tcell.ModNone), setFocus)
		root.Draw(screen)
	}
	mouse(tview.MouseLeftDown)
	mouse(tview.MouseLeftUp)
	mouse(tview.MouseLeftClick)
	assertSidebarSelection(t, sidebar, screen, 2)
}

func TestSidebarTinyResizeCannotClickInvisibleFields(t *testing.T) {
	for _, size := range [][2]int{{30, 2}, {30, 3}, {2, 10}} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			sidebar, screen := sidebarTestUI(t, 15, "one", "two", "three")
			sidebar.FocusField(1)
			sidebar.SetRect(50, 7, size[0], size[1])
			sidebar.Draw(screen)
			// The first field used to occupy this point before the resize.
			sidebarMouse(sidebar, tview.MouseLeftClick, 51, 8)
			if sidebar.GetCurrentFieldIndex() != 1 || App.GetFocus() != sidebar.Flex.GetItem(1) {
				t.Fatal("tiny viewport selected an invisible field through its stale rectangle")
			}
			for i := 0; i < sidebar.Flex.GetItemCount(); i++ {
				_, _, width, height := sidebar.Flex.GetItem(i).GetRect()
				if width > 0 && height > 0 {
					t.Fatal("undrawn field retained a clickable rectangle")
				}
			}
		})
	}
}

func TestSidebarMousePreservesEditing(t *testing.T) {
	sidebar, screen := sidebarTestUI(t, 15, "first", "second")
	field := sidebar.Flex.GetItem(1).(*tview.TextArea)
	x, y, _, _ := field.GetRect()
	sidebarMouse(sidebar, tview.MouseLeftClick, x+1, y+1)
	setFocus := func(p tview.Primitive) { App.SetFocus(p) }
	sidebar.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone), setFocus)
	if field.GetDisabled() {
		t.Fatal("edit command did not target the clicked field")
	}
	field.SetText("unsaved", false)
	sidebarMouse(sidebar, tview.MouseLeftDown, 55, 9)
	sidebarMouse(sidebar, tview.MouseLeftClick, 55, 9)
	sidebarMouse(sidebar, tview.MouseScrollDown, x+1, y+1)
	assertSidebarSelection(t, sidebar, screen, 1)
	if field.GetDisabled() || field.GetText() != "unsaved" {
		t.Fatal("mouse navigation lost the in-progress edit")
	}
	if !sidebarMouse(sidebar, tview.MouseLeftDown, x+1, y+1) {
		t.Fatal("editing field did not receive mouse input")
	}
	sidebar.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), setFocus)
	if !field.GetDisabled() || field.GetText() != "second" {
		t.Fatal("discard did not restore the clicked field")
	}
}

func TestSidebarEmptyAndShortViewport(t *testing.T) {
	sidebar, screen := sidebarTestUI(t, 4)
	sidebar.FocusFirstField()
	sidebar.FocusLastField()
	sidebar.FocusNextField()
	sidebar.FocusPreviousField()
	sidebar.FocusField(-1)
	sidebar.FocusField(10)
	sidebarMouse(sidebar, tview.MouseScrollDown, 55, 9)
	long := strings.Repeat("x", 200)
	sidebar.AddField("long", long, 28, false)
	sidebar.AddField("next", "value", 28, false)
	sidebar.FocusField(0)
	assertSidebarSelection(t, sidebar, screen, 0)
	sidebar.FocusLastField()
	assertSidebarSelection(t, sidebar, screen, 1)
	sidebar.Clear()
	sidebar.Draw(screen)
	sidebar.AddField("new", "new", 28, false)
	sidebar.FocusFirstField()
	assertSidebarSelection(t, sidebar, screen, 0)
}
