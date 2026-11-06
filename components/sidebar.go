package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/commands"
	"github.com/jorgerojas26/lazysql/lib"
	"github.com/jorgerojas26/lazysql/models"
)

type SidebarState struct {
	dbProvider        string
	currentFieldIndex int
	firstVisibleField int
}

type SidebarFieldParameters struct {
	OriginalValue string
	Height        int
}

type Sidebar struct {
	*tview.Frame
	Flex            *tview.Flex
	state           *SidebarState
	FieldParameters []*SidebarFieldParameters
	subscribers     []chan models.StateChange
	ReadOnly        bool
}

func NewSidebar(dbProvider string, readOnly bool) *Sidebar {
	flex := tview.NewFlex().SetDirection(tview.FlexColumnCSS)
	frame := tview.NewFrame(flex)
	frame.SetBackgroundColor(app.Styles.PrimitiveBackgroundColor)
	frame.SetBorder(true)
	frame.SetBorders(0, 0, 0, 0, 0, 0)

	sidebarState := &SidebarState{
		currentFieldIndex: 0,
		dbProvider:        dbProvider,
	}

	newSidebar := &Sidebar{
		Frame:       frame,
		Flex:        flex,
		state:       sidebarState,
		subscribers: []chan models.StateChange{},
		ReadOnly:    readOnly,
	}

	newSidebar.SetInputCapture(newSidebar.inputCapture)

	return newSidebar
}

func (sidebar *Sidebar) AddField(title, text string, fieldWidth int, pendingEdit bool) {
	field := tview.NewTextArea()
	field.SetWrap(true)
	field.SetDisabled(true)

	field.SetBorder(true)
	field.SetTitle(title)
	field.SetTitleAlign(tview.AlignLeft)
	field.SetTitleColor(app.Styles.PrimaryTextColor)
	field.SetText(text, true)
	field.SetTextStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor).Foreground(tview.Styles.SecondaryTextColor))

	if pendingEdit {
		sidebar.SetEditedStyles(field)
	}

	textLength := len(field.GetText())

	itemFixedSize := 3

	if textLength >= fieldWidth*3 {
		itemFixedSize = 5
	} else if textLength >= fieldWidth {
		itemFixedSize = 4
	} else {
		field.SetWrap(false)
	}

	index := sidebar.Flex.GetItemCount()
	field.SetFocusFunc(func() {
		sidebar.SetCurrentFieldIndex(index)
	})

	fieldParameters := &SidebarFieldParameters{
		Height:        itemFixedSize,
		OriginalValue: text,
	}

	sidebar.FieldParameters = append(sidebar.FieldParameters, fieldParameters)
	sidebar.Flex.AddItem(field, itemFixedSize, 0, true)
}

func (sidebar *Sidebar) Draw(screen tcell.Screen) {
	_, _, width, height := sidebar.GetInnerRect()
	if width <= 0 || height < 2 {
		// Frame skips its child at these sizes, so discard stale click targets.
		for i := 0; i < sidebar.Flex.GetItemCount(); i++ {
			sidebar.Flex.GetItem(i).SetRect(0, 0, 0, 0)
		}
		sidebar.Frame.Draw(screen)
		return
	}
	selected := sidebar.GetCurrentFieldIndex()
	first := min(sidebar.state.firstVisibleField, selected)
	used := 0
	for i := first; i <= selected && i < len(sidebar.FieldParameters); i++ {
		used += sidebar.FieldParameters[i].Height
	}
	for first < selected && used > height {
		used -= sidebar.FieldParameters[first].Height
		first++
	}
	sidebar.state.firstVisibleField = first
	remaining := height
	for i, parameters := range sidebar.FieldParameters {
		size := 0
		if i >= first {
			size = min(parameters.Height, remaining)
			remaining -= size
		}
		sidebar.Flex.ResizeItem(sidebar.Flex.GetItem(i), size, 0)
	}
	sidebar.Frame.Draw(screen)
}

func (sidebar *Sidebar) Focus(delegate func(tview.Primitive)) {
	if sidebar.Flex.GetItemCount() > 0 {
		delegate(sidebar.Flex.GetItem(sidebar.GetCurrentFieldIndex()))
	}
}

func (sidebar *Sidebar) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return sidebar.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		if !sidebar.InRect(event.Position()) || sidebar.Flex.GetItemCount() == 0 {
			return false, nil
		}
		current := sidebar.Flex.GetItem(sidebar.GetCurrentFieldIndex()).(*tview.TextArea)
		if !current.GetDisabled() {
			// Keep an in-progress edit on its field until it is committed or discarded.
			_, capture := current.MouseHandler()(action, event, setFocus)
			return true, capture
		}
		switch action {
		case tview.MouseScrollUp, tview.MouseScrollDown:
			index := sidebar.GetCurrentFieldIndex() - 1
			if action == tview.MouseScrollDown {
				index = sidebar.GetCurrentFieldIndex() + 1
			}
			index = max(0, min(index, sidebar.Flex.GetItemCount()-1))
			setFocus(sidebar.Flex.GetItem(index))
			return true, nil
		case tview.MouseLeftDown:
			// Select on release, before any focus-driven layout shift can retarget the click.
			return true, nil
		case tview.MouseLeftClick:
			for i := 0; i < sidebar.Flex.GetItemCount(); i++ {
				field := sidebar.Flex.GetItem(i).(*tview.TextArea)
				if field.InRect(event.Position()) {
					setFocus(field)
					return true, nil
				}
			}
			setFocus(sidebar)
			return true, nil
		}
		return false, nil
	})
}

func (sidebar *Sidebar) FocusNextField() {
	sidebar.FocusField(sidebar.GetCurrentFieldIndex() + 1)
}

func (sidebar *Sidebar) FocusPreviousField() {
	sidebar.FocusField(sidebar.GetCurrentFieldIndex() - 1)
}

func (sidebar *Sidebar) FocusFirstField() {
	sidebar.state.firstVisibleField = 0
	sidebar.FocusField(0)
}

func (sidebar *Sidebar) FocusLastField() {
	sidebar.FocusField(sidebar.Flex.GetItemCount() - 1)
}

func (sidebar *Sidebar) FocusField(index int) {
	if index < 0 || index >= sidebar.Flex.GetItemCount() {
		return
	}
	sidebar.SetCurrentFieldIndex(index)
	App.SetFocus(sidebar.Flex.GetItem(index))
}

func (sidebar *Sidebar) Clear() {
	sidebar.SetCurrentFieldIndex(0)
	sidebar.state.firstVisibleField = 0
	sidebar.FieldParameters = make([]*SidebarFieldParameters, 0)
	sidebar.Flex.Clear()
}

func (sidebar *Sidebar) EditTextCurrentField() {
	index := sidebar.GetCurrentFieldIndex()
	item := sidebar.Flex.GetItem(index).(*tview.TextArea)

	sidebar.SetEditingStyles(item)
}

func (sidebar *Sidebar) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	if sidebar.Flex.GetItemCount() == 0 {
		return event
	}
	command := app.Keymaps.Group(app.SidebarGroup).Resolve(event)

	switch command {
	case commands.UnfocusSidebar:
		sidebar.Publish(models.StateChange{Key: eventSidebarUnfocusing, Value: nil})
	case commands.ToggleSidebar:
		sidebar.Publish(models.StateChange{Key: eventSidebarToggling, Value: nil})
	case commands.MoveDown:
		sidebar.FocusNextField()
	case commands.MoveUp:
		sidebar.FocusPreviousField()
	case commands.GotoStart:
		sidebar.FocusFirstField()
	case commands.GotoEnd:
		sidebar.FocusLastField()
	case commands.Edit:
		if sidebar.ReadOnly {
			// no editing in read-only mode
			return event
		}

		sidebar.Publish(models.StateChange{Key: eventSidebarEditing, Value: true})

		currentItemIndex := sidebar.GetCurrentFieldIndex()
		item := sidebar.Flex.GetItem(currentItemIndex).(*tview.TextArea)
		text := item.GetText()

		columnName := item.GetTitle()
		columnNameSplit := strings.Split(columnName, "[")
		columnName = columnNameSplit[0]

		sidebar.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			command := app.Keymaps.Group(app.SidebarGroup).Resolve(event)

			switch command {
			case commands.CommitEdit:
				sidebar.SetInputCapture(sidebar.inputCapture)
				originalValue := sidebar.FieldParameters[currentItemIndex].OriginalValue
				newText := item.GetText()

				if originalValue == newText {
					sidebar.SetDisabledStyles(item)
				} else {
					sidebar.SetEditedStyles(item)
					sidebar.Publish(models.StateChange{Key: eventSidebarCommitEditing, Value: models.SidebarEditingCommitParams{ColumnName: columnName, Type: models.String, NewValue: newText}})
				}

				return nil
			case commands.DiscardEdit:
				sidebar.SetInputCapture(sidebar.inputCapture)
				sidebar.SetDisabledStyles(item)
				item.SetText(text, true)
				sidebar.Publish(models.StateChange{Key: eventSidebarEditing, Value: false})
				return nil
			}

			return event
		})

		sidebar.EditTextCurrentField()

		return nil
	case commands.SetValue:
		currentItemIndex := sidebar.GetCurrentFieldIndex()
		item := sidebar.Flex.GetItem(currentItemIndex).(*tview.TextArea)
		x, y, _, _ := item.GetRect()

		columnName := item.GetTitle()
		columnNameSplit := strings.Split(columnName, "[")
		columnName = columnNameSplit[0]

		list := NewSetValueList(sidebar.state.dbProvider)

		sidebar.Publish(models.StateChange{Key: eventSidebarEditing, Value: true})

		list.OnFinish(func(selection models.CellValueType, value string) {
			sidebar.Publish(models.StateChange{Key: eventSidebarEditing, Value: false})
			App.SetFocus(item)

			if selection >= 0 {
				sidebar.SetEditedStyles(item)
				item.SetText(value, true)
				sidebar.Publish(models.StateChange{Key: eventSidebarCommitEditing, Value: models.SidebarEditingCommitParams{ColumnName: columnName, Type: selection, NewValue: value}})
			}
		})

		list.Show(x, y, 30)

		return nil
	case commands.Copy:
		currentItemIndex := sidebar.GetCurrentFieldIndex()
		item := sidebar.Flex.GetItem(currentItemIndex).(*tview.TextArea)
		text := item.GetText()

		clipboard := lib.NewClipboard()

		err := clipboard.Write(text)
		if err != nil {
			sidebar.Publish(models.StateChange{Key: eventSidebarError, Value: err.Error()})
		}
	}
	return event
}

func (sidebar *Sidebar) SetEditingStyles(item *tview.TextArea) {
	item.SetBackgroundColor(app.Styles.SecondaryTextColor)
	item.SetTextStyle(tcell.StyleDefault.Background(app.Styles.SecondaryTextColor).Foreground(tview.Styles.ContrastSecondaryTextColor))
	item.SetTitleColor(app.Styles.ContrastSecondaryTextColor)
	item.SetBorderColor(app.Styles.SecondaryTextColor)

	item.SetWrap(true)
	item.SetDisabled(false)
}

func (sidebar *Sidebar) SetDisabledStyles(item *tview.TextArea) {
	item.SetBackgroundColor(app.Styles.PrimitiveBackgroundColor)
	item.SetTextStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor).Foreground(tview.Styles.SecondaryTextColor))
	item.SetTitleColor(app.Styles.PrimaryTextColor)
	item.SetBorderColor(app.Styles.BorderColor)

	item.SetWrap(true)
	item.SetDisabled(true)
}

func (sidebar *Sidebar) SetEditedStyles(item *tview.TextArea) {
	item.SetBackgroundColor(app.Styles.TableChangeColor)
	item.SetTextStyle(tcell.StyleDefault.Background(app.Styles.TableChangeColor).Foreground(tview.Styles.ContrastSecondaryTextColor))
	item.SetTitleColor(app.Styles.ContrastSecondaryTextColor)
	item.SetBorderColor(app.Styles.ContrastSecondaryTextColor)

	item.SetWrap(true)
	item.SetDisabled(true)
}

// Getters
func (sidebar *Sidebar) GetCurrentFieldIndex() int {
	return sidebar.state.currentFieldIndex
}

// Setters
func (sidebar *Sidebar) SetCurrentFieldIndex(index int) {
	sidebar.state.currentFieldIndex = index
}

// Subscribe to changes in the sidebar state
func (sidebar *Sidebar) Subscribe() chan models.StateChange {
	subscriber := make(chan models.StateChange)
	sidebar.subscribers = append(sidebar.subscribers, subscriber)
	return subscriber
}

// Publish subscribers of changes in the sidebar state
func (sidebar *Sidebar) Publish(change models.StateChange) {
	for _, subscriber := range sidebar.subscribers {
		subscriber <- change
	}
}
