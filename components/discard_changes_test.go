package components

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/models"
)

func newDiscardTestHome(t *testing.T) (*Home, *ResultsTable) {
	t.Helper()
	oldPages, oldFocus := mainPages, app.App.GetFocus()
	mainPages = tview.NewPages()
	t.Cleanup(func() {
		mainPages = oldPages
		app.App.SetFocus(oldFocus)
	})
	home := &Home{TabbedPane: NewTabbedPane()}
	table := newRefreshCallTable(newRefreshCallDriver())
	table.Home = home
	table.state.listOfDBChanges = &home.ListOfDBChanges
	table.Page = tview.NewPages()
	table.Page.AddPage(pageNameTable, table.Table, true, true)
	home.TabbedPane.AppendTab("orders", table, "database.orders")
	table.SetRecords([][]string{{"id"}, {"original"}, {"delete-me"}})
	home.ListOfDBChanges = []models.DBDMLChange{
		{Database: "database", Table: "orders", Type: models.DMLUpdateType, Values: []models.CellValue{{TableRowIndex: 1, TableColumnIndex: 0, Value: "edited"}}},
		{Database: "database", Table: "orders", Type: models.DMLDeleteType, Values: []models.CellValue{{TableRowIndex: 2}}},
		{Database: "database", Table: "orders", Type: models.DMLInsertType, Values: []models.CellValue{{TableRowIndex: 3, Value: "added"}}},
	}
	table.SetCell(1, 0, tview.NewTableCell("edited").SetBackgroundColor(app.Styles.TableChangeColor))
	table.SetRowColor(2, app.Styles.TableDeleteColor)
	table.SetCell(3, 0, tview.NewTableCell("added").SetBackgroundColor(app.Styles.TableInsertColor))
	mainPages.AddPage("home", home.TabbedPane, true, true)
	app.App.SetFocus(table.Table)
	return home, table
}

func openDiscardTestModal(t *testing.T, home *Home) *tview.Modal {
	t.Helper()
	if event := home.homeInputCapture(tcell.NewEventKey(tcell.KeyF8, 0, tcell.ModNone)); event != nil {
		t.Fatal("F8 was not handled")
	}
	_, primitive := mainPages.GetFrontPage()
	modal, ok := primitive.(*tview.Modal)
	if !ok {
		t.Fatalf("F8 did not show discard confirmation: %T", primitive)
	}
	return modal
}

func TestDiscardChangesCancel(t *testing.T) {
	for _, key := range []tcell.Key{tcell.KeyEnter, tcell.KeyEscape} {
		t.Run(tcell.NewEventKey(key, 0, tcell.ModNone).Name(), func(t *testing.T) {
			home, table := newDiscardTestHome(t)
			original := append([]models.DBDMLChange(nil), home.ListOfDBChanges...)
			modal := openDiscardTestModal(t, home)
			modal.InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
			if !reflect.DeepEqual(home.ListOfDBChanges, original) || table.GetCell(1, 0).Text != "edited" || table.GetRowCount() != 4 {
				t.Fatal("cancelling changed pending operations or displayed rows")
			}
			if name, _ := mainPages.GetFrontPage(); name != "home" {
				t.Fatal("confirmation did not close")
			}
			if app.App.GetFocus() != table.Table {
				t.Fatal("focus not restored")
			}
		})
	}
}

func TestDiscardChangesAllTabs(t *testing.T) {
	home, table := newDiscardTestHome(t)
	other := newRefreshCallTable(newRefreshCallDriver())
	other.SetTableName("other")
	other.Page = tview.NewPages()
	other.Page.AddPage(pageNameTable, other.Table, true, true)
	other.state.listOfDBChanges = &home.ListOfDBChanges
	other.SetRecords([][]string{{"id"}, {"other-original"}})
	other.SetCell(1, 0, tview.NewTableCell("other-edited"))
	home.TabbedPane.AppendTab("other", other, "database.other")
	home.ListOfDBChanges = append(home.ListOfDBChanges,
		models.DBDMLChange{Database: "database", Table: "other", Type: models.DMLUpdateType},
		models.DBDMLChange{Database: "database", Table: "closed", Type: models.DMLInsertType},
	)
	modal := openDiscardTestModal(t, home)
	modal.InputHandler()(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	modal.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	if len(home.ListOfDBChanges) != 0 {
		t.Fatal("pending changes remain")
	}
	if table.GetRowCount() != 3 || table.GetCell(1, 0).Text != "original" || table.GetCell(2, 0).Text != "delete-me" {
		t.Fatal("edits, insertions, or deletions not restored")
	}
	if other.GetCell(1, 0).Text != "other-original" {
		t.Fatal("other tab not restored")
	}
	for _, row := range []int{1, 2} {
		color := table.GetCell(row, 0).BackgroundColor
		if color == app.Styles.TableDeleteColor || color == app.Styles.TableChangeColor {
			t.Fatal("pending highlight remains")
		}
	}
	if name, _ := mainPages.GetFrontPage(); name != "home" {
		t.Fatal("confirmation did not close")
	}
}

func TestDiscardChangesFilteredTabsAndIdentityDrift(t *testing.T) {
	home, table := newDiscardTestHome(t)
	table.SetDatabaseName("other-database")
	home.ListOfDBChanges[0].Database = "other-database"
	filtered := newRefreshCallTable(newRefreshCallDriver())
	filtered.Page = tview.NewPages()
	filtered.Page.AddPage(pageNameTable, filtered.Table, true, true)
	filtered.SetRecords([][]string{{"id"}, {"filtered-original"}})
	filtered.SetCell(1, 0, tview.NewTableCell("filtered-edited"))
	home.TabbedPane.AppendTab("filtered", filtered, "database.orders#2")
	home.discardPendingChanges()
	if table.GetCell(1, 0).Text != "original" || filtered.GetCell(1, 0).Text != "filtered-original" {
		t.Fatal("discard missed a filtered tab or a changed table identity")
	}
}

func TestDiscardChangesRestoresSidebarFocus(t *testing.T) {
	home, table := newDiscardTestHome(t)
	table.Sidebar = NewSidebar("mysql", false)
	table.SidebarContainer = tview.NewFlex()
	table.state.showSidebar = true
	table.SetColumns([][]string{{"name", "type"}, {"id", "text"}})
	table.Sidebar.AddField("id", "edited", 30, true)
	oldField := table.Sidebar.Flex.GetItem(0)
	app.App.SetFocus(oldField)
	modal := openDiscardTestModal(t, home)
	modal.InputHandler()(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	modal.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	if !table.Sidebar.HasFocus() || app.App.GetFocus() == oldField {
		t.Fatal("focus restored to an orphaned sidebar field")
	}
}

func TestDiscardChangesRecoversStolenFocus(t *testing.T) {
	home, table := newDiscardTestHome(t)
	openDiscardTestModal(t, home)
	app.App.SetFocus(table.Table) // A load finishing can move focus behind the modal.
	mainPages.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	if name, _ := mainPages.GetFrontPage(); name != "home" || len(home.ListOfDBChanges) != 3 {
		t.Fatal("discard confirmation cannot be cancelled after background focus change")
	}
}

func TestDiscardChangesPreservesOtherSurfaces(t *testing.T) {
	home, _ := newDiscardTestHome(t)
	metadata := newRefreshCallTable(newRefreshCallDriver())
	metadata.SetTableName("metadata")
	metadata.Menu.SetSelectedOption(2)
	metadata.Page = tview.NewPages()
	metadata.Page.AddPage(pageNameTable, metadata.Table, true, true)
	metadata.SetCell(0, 0, tview.NewTableCell("column_name"))
	metadata.SetCell(1, 0, tview.NewTableCell("metadata-unchanged"))
	home.TabbedPane.AppendTab("metadata", metadata, "database.metadata")
	home.ListOfDBChanges = append(home.ListOfDBChanges, models.DBDMLChange{Database: "database", Table: "metadata", Type: models.DMLUpdateType})

	editor := newRefreshCallTable(newRefreshCallDriver())
	editor.Editor = NewSQLEditor("")
	editor.Editor.SetText("select 'unsaved SQL';", true)
	editor.Page = tview.NewPages()
	editor.Page.AddPage(pageNameTable, editor.Editor, true, true)
	home.TabbedPane.AppendTab("editor", editor, "editor")
	app.App.SetFocus(editor.Editor)

	modal := openDiscardTestModal(t, home)
	modal.InputHandler()(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	modal.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	if len(home.ListOfDBChanges) != 0 {
		t.Fatal("pending changes remain")
	}
	if editor.Editor.GetText() != "select 'unsaved SQL';" || app.App.GetFocus() != editor.Editor {
		t.Fatal("discard changed SQL editor content or focus")
	}
	if metadata.GetCell(1, 0).Text != "metadata-unchanged" || metadata.Menu.GetSelectedOption() != 2 {
		t.Fatal("discard replaced the metadata surface")
	}
}

func TestDiscardChangesIgnoresEmptyAndEditing(t *testing.T) {
	for _, mode := range []string{"empty", "editing", "filtering"} {
		t.Run(mode, func(t *testing.T) {
			home, table := newDiscardTestHome(t)
			switch mode {
			case "empty":
				home.ListOfDBChanges = nil
			case "editing":
				table.SetIsEditing(true)
			case "filtering":
				table.SetIsFiltering(true)
			}
			home.homeInputCapture(tcell.NewEventKey(tcell.KeyF8, 0, tcell.ModNone))
			if name, _ := mainPages.GetFrontPage(); name != "home" {
				t.Fatal("unexpected discard confirmation")
			}
		})
	}
}
