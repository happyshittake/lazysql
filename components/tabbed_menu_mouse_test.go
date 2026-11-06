package components

import (
	"testing"

	"github.com/rivo/tview"
)

func appendMouseTestTable(home *Home, name string) *ResultsTable {
	table := newRecordsFetchTestTable(&schemaProgrammingMock{})
	table.Home, table.Tree = home, home.Tree
	table.Wrapper, table.SidebarContainer = tview.NewFlex().SetDirection(tview.FlexColumnCSS), tview.NewFlex()
	table.Page = tview.NewPages().AddPage(pageNameTable, table.Wrapper, true, true)
	table.Sidebar = NewSidebar("mock", false)
	table.WithFilter()
	table.state.records = [][]string{{"id"}, {name}}
	table.UpdateRows(table.GetRecords())
	home.TabbedPane.AppendTab(name, table, name)
	return table
}

func TestUpperTabMouseSwitchesPageAndFocus(t *testing.T) {
	home, first, screen := tableDetailMouseUI(t, false)
	second := appendMouseTestTable(home, "orders")
	home.focusLeftWrapper()
	home.Draw(screen)
	firstTab := home.TabbedPane.GetTabByReference("table")
	x, y, _, _ := firstTab.Header.GetRect()
	clickTableDetail(home, screen, x+1, y)
	if got, _ := home.TabbedPane.GetFrontPage(); got != "table" || home.TabbedPane.GetCurrentTab() != firstTab {
		t.Fatalf("upper tab click left page %q selected instead of table", got)
	}
	if !first.HasFocus() {
		t.Fatal("upper tab click did not focus selected table")
	}
	assertDetailPaneFocus(t, home, focusedWrapperRight)
	x, y, _, _ = home.TabbedPane.GetTabByReference("orders").Header.GetRect()
	clickTableDetail(home, screen, x+1, y)
	if got, _ := home.TabbedPane.GetFrontPage(); got != "orders" || !second.HasFocus() {
		t.Fatal("clicking back did not select and focus orders")
	}
}

func TestUpperTabMousePreservesSidebarEdit(t *testing.T) {
	home, first, screen := tableDetailMouseUI(t, false)
	appendMouseTestTable(home, "orders")
	home.TabbedPane.SetCurrentTab(home.TabbedPane.GetTabByReference("table"))
	addMouseTestSidebar(home, first, screen, false)
	field := first.Sidebar.Flex.GetItem(0).(*tview.TextArea)
	first.Sidebar.SetEditingStyles(field)
	first.SetIsEditing(true)
	field.SetText("draft", false)
	App.SetFocus(field)
	x, y, _, _ := home.TabbedPane.GetTabByReference("orders").Header.GetRect()
	clickTableDetail(home, screen, x+1, y)
	if got, _ := home.TabbedPane.GetFrontPage(); got != "table" || App.GetFocus() != field || field.GetText() != "draft" {
		t.Fatal("tab click interrupted an in-progress sidebar edit")
	}
}

func TestUpperTabMouseAllowsEditorSwitchWithoutLosingSQL(t *testing.T) {
	home, first, screen := tableDetailMouseUI(t, false)
	first.WithEditor()
	first.Editor.SetText("select 42", false)
	second := appendMouseTestTable(home, "orders")
	home.TabbedPane.SetCurrentTab(home.TabbedPane.GetTabByReference("table"))
	App.SetFocus(first.Editor)
	home.Draw(screen)
	x, y, _, _ := home.TabbedPane.GetTabByReference("orders").Header.GetRect()
	clickTableDetail(home, screen, x+1, y)
	if !second.HasFocus() {
		t.Fatal("SQL editor editing state prevented tab switching")
	}
	x, y, _, _ = home.TabbedPane.GetTabByReference("table").Header.GetRect()
	clickTableDetail(home, screen, x+1, y)
	if !first.Editor.HasFocus() || first.Editor.GetText() != "select 42" {
		t.Fatal("returning to editor lost SQL or editor focus")
	}
}

func TestUpperTabMouseIgnoresHiddenHeaderRectangles(t *testing.T) {
	home, _, screen := tableDetailMouseUI(t, false)
	appendMouseTestTable(home, "orders")
	appendMouseTestTable(home, "products")
	appendMouseTestTable(home, "invoices")
	appendMouseTestTable(home, "sessions")
	home.Draw(screen) // Every header has a rectangle from the wide layout.
	home.SetRect(0, 0, 50, 45)
	home.Draw(screen) // Earlier headers are hidden; their old rectangles can overlap visible tabs.
	selected := home.TabbedPane.GetCurrentTab()
	x, y, _, _ := selected.Header.GetRect()
	clickTableDetail(home, screen, x+1, y)
	if home.TabbedPane.GetCurrentTab() != selected {
		t.Fatal("a hidden header intercepted a visible header click after resize")
	}
	// Undrawn headers must not retain hit targets from a previous layout.
	hidden := home.TabbedPane.GetTabByReference("table")
	_, _, width, height := hidden.Header.GetRect()
	if width > 0 && height > 0 {
		t.Fatal("hidden header retained a stale clickable rectangle")
	}
}
