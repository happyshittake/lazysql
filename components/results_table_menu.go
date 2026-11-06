package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/app"
)

type ResultsTableMenuState struct {
	SelectedOption int
}

type ResultsTableMenu struct {
	*tview.Flex
	state      *ResultsTableMenuState
	MenuItems  []*tview.TextView
	onSelected func(int)
}

var menuItems = []string{
	menuRecords,
	menuColumns,
	menuConstraints,
	menuForeignKeys,
	menuIndexes,
}

func NewResultsTableMenu() *ResultsTableMenu {
	state := &ResultsTableMenuState{
		SelectedOption: 1,
	}

	menu := &ResultsTableMenu{
		Flex:  tview.NewFlex(),
		state: state,
	}

	menu.SetBorder(true)

	for i, item := range menuItems {
		separator := " | "
		if i == len(menuItems)-1 {
			separator = ""
		}

		text := fmt.Sprintf("%s [%d] %s", item, i+1, separator)
		textview := tview.NewTextView().SetText(text)

		if i == 0 {
			textview.SetTextColor(app.Styles.PrimaryTextColor)
		}

		size := 15

		switch item {
		case menuConstraints:
			size = 19
		case menuForeignKeys:
			size = 20
		case menuIndexes:
			size = 16
		}

		menu.MenuItems = append(menu.MenuItems, textview)
		menu.AddItem(textview, size, 0, false)
	}

	return menu
}

func (menu *ResultsTableMenu) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return menu.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		x, y := event.Position()
		innerX, innerY, width, height := menu.GetInnerRect()
		if x < innerX || x >= innerX+width || y < innerY || y >= innerY+height {
			return false, nil
		}
		if action != tview.MouseLeftDown && action != tview.MouseLeftClick {
			return false, nil
		}
		for i, item := range menu.MenuItems {
			itemX, _, _, _ := item.GetRect()
			labelWidth := tview.TaggedStringWidth(fmt.Sprintf("%s [%d]", menuItems[i], i+1))
			if item.InRect(x, y) && x < itemX+labelWidth {
				if action == tview.MouseLeftClick && menu.onSelected != nil {
					menu.onSelected(i + 1)
				}
				return true, nil
			}
		}
		return false, nil
	})
}

// Getters and Setters
func (menu *ResultsTableMenu) GetSelectedOption() int {
	return menu.state.SelectedOption
}

func (menu *ResultsTableMenu) SetSelectedOption(option int) {
	if menu.state.SelectedOption != option {

		menu.state.SelectedOption = option

		itemCount := menu.GetItemCount()

		for i := 0; i < itemCount; i++ {
			menu.GetItem(i).(*tview.TextView).SetTextColor(app.Styles.PrimaryTextColor)
		}

		menu.GetItem(option - 1).(*tview.TextView).SetTextColor(app.Styles.SecondaryTextColor)
	}
}

func (menu *ResultsTableMenu) SetBlur() {
	menu.SetBorderColor(app.Styles.InverseTextColor)

	for _, item := range menu.MenuItems {
		item.SetTextColor(app.Styles.InverseTextColor)
	}
}

func (menu *ResultsTableMenu) SetFocus() {
	menu.SetBorderColor(app.Styles.PrimaryTextColor)

	for i, item := range menu.MenuItems {
		if i+1 == menu.GetSelectedOption() {
			item.SetTextColor(app.Styles.SecondaryTextColor)
		} else {
			item.SetTextColor(app.Styles.PrimaryTextColor)
		}
	}
}
