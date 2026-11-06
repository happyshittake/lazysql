package components

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/app"
)

func tableDetailMouseUI(t *testing.T, overlay bool) (*Home, *ResultsTable, tcell.Screen) {
	t.Helper()
	oldApplication, oldOverlay, oldPages := App.Application, App.Config().SidebarOverlay, mainPages
	App.Application = tview.NewApplication()
	App.Config().SidebarOverlay = overlay
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(160, 45)
	App.SetScreen(screen)
	t.Cleanup(func() {
		screen.Fini()
		App.Application = oldApplication
		App.Config().SidebarOverlay = oldOverlay
		mainPages = oldPages
	})
	driver := &schemaProgrammingMock{}
	tree := NewTree("", driver, nil)
	tree.SetFocusFunc(nil) // No database initialization is needed to test mouse routing.
	home := &Home{
		Flex: tview.NewFlex(), Tree: tree, TabbedPane: NewTabbedPane(),
		LeftWrapper: tview.NewFlex(), RightWrapper: tview.NewFlex(),
		leftWrapperVisible: true, treePinned: true, HelpStatus: NewHelpStatus(),
	}
	table := newRecordsFetchTestTable(driver)
	table.Home, table.Tree = home, tree
	table.Wrapper, table.SidebarContainer = tview.NewFlex().SetDirection(tview.FlexColumnCSS), tview.NewFlex()
	table.Page = tview.NewPages().AddPage(pageNameTable, table.Wrapper, true, true)
	table.Sidebar = NewSidebar("mock", false)
	table.Page.AddPage(pageNameSidebar, table.Sidebar, false, false)
	table.WithFilter()
	table.state.records = [][]string{{"id"}, {"record"}}
	table.state.columns = [][]string{{"Field", "Type"}, {"id", "int"}}
	table.state.constraints = [][]string{{"constraint"}, {"constraint-row"}}
	table.state.foreignKeys = [][]string{{"foreign-key"}, {"foreign-key-row"}}
	table.state.indexes = [][]string{{"index"}, {"index-row"}}
	table.UpdateRows(table.GetRecords())
	table.SetSelectable(true, true)
	table.SetFixed(1, 0)
	table.SetInputCapture(table.tableInputCapture)
	home.TabbedPane.AppendTab("table", table, "table")
	home.LeftWrapper.AddItem(tree.Wrapper, 0, 1, true)
	home.RightWrapper.SetBorder(true)
	home.RightWrapper.SetDirection(tview.FlexColumnCSS).
		AddItem(home.TabbedPane.HeaderContainer, 1, 0, false).
		AddItem(home.TabbedPane.Pages, 0, 1, false)
	home.AddItem(home.LeftWrapper, 25, 0, false).AddItem(home.RightWrapper, 0, 1, false)
	home.SetInputCapture(home.homeInputCapture)
	home.SetRect(0, 0, 160, 45)
	mainPages = tview.NewPages().AddPage("home", home, true, true)
	App.SetRoot(home, false)
	home.focusLeftWrapper()
	home.Draw(screen)
	return home, table, screen
}

func clickTableDetail(home *Home, screen tcell.Screen, x, y int) {
	for _, action := range []tview.MouseAction{tview.MouseLeftDown, tview.MouseLeftUp, tview.MouseLeftClick} {
		home.MouseHandler()(action, tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
		home.Draw(screen)
	}
}

func assertDetailPaneFocus(t *testing.T, home *Home, want string) {
	t.Helper()
	right, left := app.Styles.InverseTextColor, app.Styles.PrimaryTextColor
	if want == focusedWrapperRight {
		right, left = left, right
	}
	if home.FocusedWrapper != want || home.RightWrapper.GetBorderColor() != right || home.LeftWrapper.GetBorderColor() != left {
		t.Fatalf("pane focus = %q, borders left=%v right=%v; want %q with synchronized colors", home.FocusedWrapper, home.LeftWrapper.GetBorderColor(), home.RightWrapper.GetBorderColor(), want)
	}
	table := home.TabbedPane.GetCurrentTab().Content.(*ResultsTable)
	if table.Menu.GetBorderColor() != right || table.GetBorderColor() != right {
		t.Fatal("table and menu retained the wrong focus styling")
	}
}

func TestTableDetailMouseFocusBothDirections(t *testing.T) {
	home, table, screen := tableDetailMouseUI(t, false)
	x, y, _, _ := table.GetInnerRect()
	clickTableDetail(home, screen, x+2, y+1)
	if !table.HasFocus() {
		t.Fatal("table did not receive focus")
	}
	assertDetailPaneFocus(t, home, focusedWrapperRight)
	x, y, _, _ = home.Tree.GetInnerRect()
	clickTableDetail(home, screen, x+2, y+2)
	if !home.Tree.HasFocus() {
		t.Fatal("tree did not receive focus")
	}
	assertDetailPaneFocus(t, home, focusedWrapperLeft)
	clickTableDetail(home, screen, 159, 44) // Right pane border.
	if !table.HasFocus() {
		t.Fatal("clicking detail border did not focus table")
	}
	assertDetailPaneFocus(t, home, focusedWrapperRight)
}

func TestTableDetailMouseFilterKeepsInputFocus(t *testing.T) {
	home, table, screen := tableDetailMouseUI(t, false)
	x, y, _, _ := table.Filter.Input.GetInnerRect()
	clickTableDetail(home, screen, x+2, y)
	if !table.Filter.Input.HasFocus() {
		t.Fatal("clicking WHERE did not retain input focus")
	}
	assertDetailPaneFocus(t, home, focusedWrapperRight)
	table.Filter.Input.InputHandler()(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
	if table.Filter.Input.GetText() != "id" {
		t.Fatal("WHERE autocomplete stopped working")
	}
}

func TestTableDetailMouseSidebarFocusBothDirections(t *testing.T) {
	for _, overlay := range []bool{false, true} {
		t.Run(fmt.Sprintf("overlay=%t", overlay), func(t *testing.T) {
			home, table, screen := tableDetailMouseUI(t, overlay)
			table.Sidebar.AddField("id", "record", 24, false)
			if overlay {
				table.Sidebar.SetRect(125, 4, 30, 30)
				table.Page.ShowPage(pageNameSidebar)
			} else {
				table.SidebarContainer.AddItem(table.Sidebar, 0, 1, true)
			}
			home.Draw(screen)
			field := table.Sidebar.Flex.GetItem(0)
			x, y, _, _ := field.GetRect()
			clickTableDetail(home, screen, x+1, y+1)
			if App.GetFocus() != field {
				t.Fatal("sidebar field did not receive focus")
			}
			assertDetailPaneFocus(t, home, focusedWrapperRight)
			x, y, _, _ = home.Tree.GetInnerRect()
			clickTableDetail(home, screen, x+2, y+2)
			assertDetailPaneFocus(t, home, focusedWrapperLeft)
			x, y, _, _ = table.GetInnerRect()
			clickTableDetail(home, screen, x+2, y+1)
			if !table.HasFocus() {
				t.Fatal("sidebar-to-table click did not focus table")
			}
			assertDetailPaneFocus(t, home, focusedWrapperRight)
		})
	}
}

func TestResultsTableMenuMouseSwitchesDisplayedSurface(t *testing.T) {
	home, table, screen := tableDetailMouseUI(t, false)
	for _, example := range []struct {
		option int
		want   string
	}{
		{2, "id"}, {3, "constraint-row"}, {4, "foreign-key-row"}, {5, "index-row"}, {1, "record"},
	} {
		x, y, _, _ := table.Menu.MenuItems[example.option-1].GetRect()
		clickTableDetail(home, screen, x+1, y)
		if table.Menu.GetSelectedOption() != example.option || table.GetCell(1, 0).Text != example.want {
			t.Fatalf("clicked option %d: selected=%d row=%q, want %q", example.option, table.Menu.GetSelectedOption(), table.GetCell(1, 0).Text, example.want)
		}
		if !table.HasFocus() {
			t.Fatal("menu click should leave keyboard focus on table")
		}
		assertDetailPaneFocus(t, home, focusedWrapperRight)
	}
}

func TestTableDetailMouseEditorKeepsInputFocus(t *testing.T) {
	home, table, screen := tableDetailMouseUI(t, false)
	table.WithEditor()
	home.Draw(screen)
	x, y, _, _ := table.Editor.GetInnerRect()
	clickTableDetail(home, screen, x+2, y+1)
	if !table.Editor.HasFocus() || !table.GetIsEditing() {
		t.Fatal("editor click did not preserve editor focus and editing state")
	}
	assertDetailPaneFocus(t, home, focusedWrapperRight)
}

func TestTableDetailMousePreservesKeyboardNavigation(t *testing.T) {
	home, table, _ := tableDetailMouseUI(t, false)
	home.homeInputCapture(tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone))
	if !table.HasFocus() {
		t.Fatal("L did not focus table")
	}
	assertDetailPaneFocus(t, home, focusedWrapperRight)
	for _, example := range []struct {
		key    rune
		option int
		want   string
	}{
		{'2', 2, "id"}, {'3', 3, "constraint-row"}, {'4', 4, "foreign-key-row"}, {'5', 5, "index-row"}, {'1', 1, "record"},
	} {
		table.InputHandler()(tcell.NewEventKey(tcell.KeyRune, example.key, tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
		if table.Menu.GetSelectedOption() != example.option || table.GetCell(1, 0).Text != example.want {
			t.Fatalf("keyboard surface switching failed for %c", example.key)
		}
	}
	home.homeInputCapture(tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModNone))
	if !home.Tree.HasFocus() {
		t.Fatal("H did not focus tree")
	}
	assertDetailPaneFocus(t, home, focusedWrapperLeft)
}

func TestResultsTableMenuMouseIgnoresClippedLabels(t *testing.T) {
	home, table, screen := tableDetailMouseUI(t, false)
	home.SetRect(0, 0, 60, 45)
	home.Draw(screen)
	x, y, width, _ := table.Menu.GetRect()
	// Columns is partially visible, but the right border is not part of its label.
	clickTableDetail(home, screen, x+width-1, y+1)
	if table.Menu.GetSelectedOption() != 1 {
		t.Fatal("border over a clipped label switched surfaces")
	}
	// Previously drawn Foreign Keys is now outside the detail pane.
	x, y, _, _ = table.Menu.MenuItems[3].GetRect()
	table.Menu.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(x+1, y, tcell.ButtonNone, tcell.ModNone), func(p tview.Primitive) { App.SetFocus(p) })
	if table.Menu.GetSelectedOption() != 1 {
		t.Fatal("hidden menu label switched surfaces")
	}
}

func TestResultsTableMenuMouseIgnoresBorderAndOutside(t *testing.T) {
	home, table, screen := tableDetailMouseUI(t, false)
	x, y, width, height := table.Menu.GetRect()
	for _, point := range [][2]int{{x, y}, {x + 20, y + height - 1}, {x + width - 2, y + 1}, {x - 1, y + 1}, {x + 13, y + 1}} {
		clickTableDetail(home, screen, point[0], point[1])
		if table.Menu.GetSelectedOption() != 1 || table.GetCell(1, 0).Text != "record" {
			t.Fatal("non-label click switched surfaces")
		}
	}
}
