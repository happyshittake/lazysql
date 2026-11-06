package components

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/drivers"
)

func addMouseTestSidebar(home *Home, table *ResultsTable, screen tcell.Screen, overlay bool) {
	table.Sidebar.AddField("id", "record", 24, false)
	table.Sidebar.AddField("name", "second", 24, false)
	if overlay {
		table.Sidebar.SetRect(125, 4, 30, 30)
		table.Page.ShowPage(pageNameSidebar)
	} else {
		table.SidebarContainer.AddItem(table.Sidebar, 0, 1, true)
	}
	home.Draw(screen)
}

func TestResultsTableMenuMousePreservesSidebarDraft(t *testing.T) {
	for _, overlay := range []bool{false, true} {
		t.Run(fmt.Sprintf("overlay=%t", overlay), func(t *testing.T) {
			home, table, screen := tableDetailMouseUI(t, overlay)
			addMouseTestSidebar(home, table, screen, overlay)
			field := table.Sidebar.Flex.GetItem(0).(*tview.TextArea)
			table.Sidebar.SetEditingStyles(field)
			table.SetIsEditing(true)
			field.SetText("unsaved draft", false)
			App.SetFocus(field)
			x, y, _, _ := table.Menu.MenuItems[1].GetRect()
			clickTableDetail(home, screen, x+1, y)
			if table.Menu.GetSelectedOption() != 1 || App.GetFocus() != field || field.GetText() != "unsaved draft" || field.GetDisabled() {
				t.Fatal("menu click changed surfaces or interrupted the sidebar draft")
			}
		})
	}
}

func TestTableDetailMouseSidebarWhitespaceReceivesKeyboardFocus(t *testing.T) {
	for _, overlay := range []bool{false, true} {
		for _, border := range []bool{false, true} {
			t.Run(fmt.Sprintf("overlay=%t/border=%t", overlay, border), func(t *testing.T) {
				home, table, screen := tableDetailMouseUI(t, overlay)
				addMouseTestSidebar(home, table, screen, overlay)
				x, y, _, height := table.Sidebar.GetRect()
				if !border {
					x++
					y += height - 2
				}
				clickTableDetail(home, screen, x, y)
				if !table.Sidebar.HasFocus() {
					t.Fatal("sidebar click highlighted right pane but left keyboard focus elsewhere")
				}
				assertDetailPaneFocus(t, home, focusedWrapperRight)
				home.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
				if table.Sidebar.GetCurrentFieldIndex() != 1 {
					t.Fatal("keyboard navigation did not target the clicked sidebar")
				}
			})
		}
	}
}

func TestTableDetailMouseDoesNotStealModalFocus(t *testing.T) {
	home, table, screen := tableDetailMouseUI(t, false)
	modal := NewConfirmationModal("Exit LazySQL?")
	dismissed := false
	modal.SetDoneFunc(func(_ int, _ string) { dismissed = true })
	mainPages.AddPage("confirmation", modal, true, true)
	mainPages.SetRect(0, 0, 160, 45)
	App.SetFocus(modal)
	mainPages.Draw(screen)
	for _, point := range [][2]int{{159, 44}, {28, 10}, {1, 1}} {
		for _, action := range []tview.MouseAction{tview.MouseLeftDown, tview.MouseLeftUp, tview.MouseLeftClick} {
			mainPages.MouseHandler()(action, tcell.NewEventMouse(point[0], point[1], tcell.ButtonNone, tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
		}
		if !modal.HasFocus() || table.HasFocus() {
			t.Fatal("background click stole focus from visible confirmation")
		}
	}
	mainPages.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
	if !dismissed {
		t.Fatal("confirmation no longer receives keyboard answers")
	}
	assertDetailPaneFocus(t, home, focusedWrapperLeft)
}

func TestTableDetailMouseArgumentConnectionRegistersHome(t *testing.T) {
	_, table, screen := tableDetailMouseUI(t, false)
	oldEngine, oldReadOnly := initialTelemetryEngine, initialTelemetryReadOnly
	t.Cleanup(func() { initialTelemetryEngine, initialTelemetryReadOnly = oldEngine, oldReadOnly })
	mainPages = tview.NewPages()
	path := filepath.Join(t.TempDir(), "mouse.sqlite3")
	db := &drivers.SQLite{}
	if err := db.Connect(App.Context(), path); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Connection.Exec("CREATE TABLE mouse (id INTEGER)"); err != nil {
		t.Fatal(err)
	}
	if err := db.Connection.Close(); err != nil {
		t.Fatal(err)
	}
	if err := InitFromArg(path, false); err != nil {
		t.Fatal(err)
	}
	_, page := mainPages.GetFrontPage()
	home, ok := page.(*Home)
	if !ok {
		t.Fatalf("argument connection registered %T, bypassing Home mouse handling", page)
	}
	t.Cleanup(func() { home.DBDriver.(*drivers.SQLite).Connection.Close() })
	home.Tree.SetFocusFunc(nil)
	home.TabbedPane.AppendTab("table", table, "table")
	home.SetRect(0, 0, 160, 45)
	home.focusLeftWrapper()
	mainPages.SetRect(0, 0, 160, 45)
	mainPages.Draw(screen)
	x, y, _, _ := table.GetInnerRect()
	for _, action := range []tview.MouseAction{tview.MouseLeftDown, tview.MouseLeftUp, tview.MouseLeftClick} {
		mainPages.MouseHandler()(action, tcell.NewEventMouse(x+2, y+1, tcell.ButtonNone, tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
	}
	if !table.HasFocus() {
		t.Fatal("argument connection table click did not focus table")
	}
	assertDetailPaneFocus(t, home, focusedWrapperRight)
}
