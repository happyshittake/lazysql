package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/app"
)

func (home *Home) showDiscardChangesConfirmation() {
	if len(home.ListOfDBChanges) == 0 || mainPages == nil || mainPages.HasPage(pageNameDiscardChanges) || quitConfirmationVisible() {
		return
	}
	previousFocus := app.App.GetFocus()
	var focusedSidebar *Sidebar
	if tab := home.TabbedPane.GetCurrentTab(); tab != nil {
		if table, ok := tab.Content.(*ResultsTable); ok && table.Sidebar != nil && table.Sidebar.HasFocus() {
			focusedSidebar = table.Sidebar
		}
	}
	previousCapture := mainPages.GetInputCapture()
	modal := tview.NewModal().
		SetText("Discard all pending edits, additions, and deletions across this connection? This cannot be undone.").
		AddButtons([]string{"Cancel", "Discard"})
	modal.SetDoneFunc(func(_ int, label string) {
		mainPages.SetInputCapture(previousCapture)
		mainPages.RemovePage(pageNameDiscardChanges)
		if label == "Discard" {
			home.discardPendingChanges()
			if focusedSidebar != nil {
				previousFocus = focusedSidebar
			}
		}
		app.App.SetFocus(previousFocus)
		keepQuitConfirmationFocused()
	})
	// Background loads can move focus; keep keyboard events on the visible dialog.
	mainPages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		app.App.SetFocus(modal)
		return event
	})
	mainPages.AddPage(pageNameDiscardChanges, modal, true, true)
	app.App.SetFocus(modal)
}

func (home *Home) discardPendingChanges() {
	changes := home.ListOfDBChanges
	home.ListOfDBChanges = nil
	affected := make(map[[2]string]bool, len(changes))
	for _, change := range changes {
		affected[[2]string{change.Database, change.Table}] = true
	}
	for tab := home.TabbedPane.state.FirstTab; tab != nil; tab = tab.NextTab {
		table, ok := tab.Content.(*ResultsTable)
		if !ok || table.Editor != nil || !affected[[2]string{table.GetDatabaseName(), table.GetTableName()}] {
			continue
		}
		if table.Menu == nil || table.Menu.GetSelectedOption() == 1 {
			// Pending operations change displayed cells, not the cached database rows.
			table.UpdateRows(table.GetRecords())
			if table.GetShowSidebar() {
				table.UpdateSidebar()
			}
		}
	}
	if home.Tree != nil {
		home.Tree.ForceRemoveHighlight()
	}
}
