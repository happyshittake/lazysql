package components

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestWhereFilterAutocompleteOnFocus(t *testing.T) {
	for _, focus := range []string{"click", "keyboard"} {
		for _, example := range []struct {
			text string
			want string
		}{
			{"", "id"},
			{"na", "name"},
			{"name is ", "name is not"},
		} {
			t.Run(focus+"/"+example.text, func(t *testing.T) {
				oldFocus := App.GetFocus()
				t.Cleanup(func() { App.SetFocus(oldFocus) })
				table := newMarkTestTable(nil)
				table.Page = tview.NewPages()
				table.Wrapper = tview.NewFlex()
				table.SidebarContainer = tview.NewFlex()
				table.Pagination = NewPagination()
				table.WithFilter()
				// Metadata arrives after the filter is constructed.
				table.state.columns = [][]string{{"Field"}, {"id"}, {"name"}}

				input := table.Filter.Input
				input.SetText(example.text)
				input.SetRect(0, 0, 40, 1)
				screen := tcell.NewSimulationScreen("")
				if err := screen.Init(); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(screen.Fini)
				input.Draw(screen)
				if focus == "click" {
					input.MouseHandler()(tview.MouseLeftDown,
						tcell.NewEventMouse(30, 0, tcell.Button1, tcell.ModNone),
						func(p tview.Primitive) { App.SetFocus(p) })
				} else {
					table.search()
				}
				if !input.HasFocus() {
					t.Fatal("WHERE input did not receive focus")
				}
				input.InputHandler()(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone),
					func(p tview.Primitive) { App.SetFocus(p) })
				if got := input.GetText(); got != example.want {
					t.Fatalf("autocomplete = %q, want %q", got, example.want)
				}
			})
		}
	}
}

func TestClickedWhereFilterBackspaceEditsWithoutNavigating(t *testing.T) {
	for _, key := range []tcell.Key{tcell.KeyBackspace, tcell.KeyBackspace2} {
		t.Run(tcell.KeyNames[key], func(t *testing.T) {
			oldPages := mainPages
			oldFocus := App.GetFocus()
			t.Cleanup(func() {
				mainPages = oldPages
				App.SetFocus(oldFocus)
			})

			table := newMarkTestTable(nil)
			table.Page = tview.NewPages()
			table.Wrapper = tview.NewFlex()
			table.SidebarContainer = tview.NewFlex()
			table.Pagination = NewPagination()
			table.WithFilter()

			home := &Home{TabbedPane: NewTabbedPane()}
			home.TabbedPane.AppendTab("customers", table, "customers")
			mainPages = tview.NewPages()
			mainPages.AddPage(pageNameConnections, tview.NewBox(), true, false)
			mainPages.AddPage("home", tview.NewBox(), true, true)

			input := table.Filter.Input
			input.SetText("name = 'abc'")
			input.SetRect(0, 0, 40, 1)
			screen := tcell.NewSimulationScreen("")
			if err := screen.Init(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(screen.Fini)
			input.Draw(screen)
			input.MouseHandler()(tview.MouseLeftDown,
				tcell.NewEventMouse(30, 0, tcell.Button1, tcell.ModNone),
				func(p tview.Primitive) { App.SetFocus(p) })
			if !input.HasFocus() {
				t.Fatal("click did not focus WHERE input")
			}

			event := tcell.NewEventKey(key, 0, tcell.ModNone)
			if forwarded := home.homeInputCapture(event); forwarded != nil {
				input.InputHandler()(forwarded, func(p tview.Primitive) { App.SetFocus(p) })
			}
			if page, _ := mainPages.GetFrontPage(); page != "home" {
				t.Fatalf("Backspace in WHERE filter navigated to %q", page)
			}
			if got := input.GetText(); got != "name = 'abc" {
				t.Fatalf("WHERE text = %q, want %q", got, "name = 'abc")
			}

			App.SetFocus(table)
			if table.GetIsFiltering() {
				t.Fatal("WHERE filter remained active after focus moved to the table")
			}
			home.homeInputCapture(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
			if page, _ := mainPages.GetFrontPage(); page != pageNameConnections {
				t.Fatalf("Backspace outside WHERE filter stayed on %q", page)
			}
		})
	}
}
