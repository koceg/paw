package ui

import (
	"context"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lucor.dev/paw/internal/azure"
	"lucor.dev/paw/internal/icon"
	"lucor.dev/paw/internal/paw"
)

type vaultView struct {
	widget.BaseWidget

	cancelCtx context.CancelFunc

	mainView *mainView

	name          *widget.Label
	vault         *azure.SecretsVault
	filterOptions *paw.VaultFilterOptions

	// view is a container used to split the vault view in two areas: navbar and content.
	// The navbar area allows to switch between the vault's item along with the possibility to filter by name, type and add new items.
	// The content area contains the views that allow to perform action on the item (i.e. read, edit, delete)
	view *fyne.Container

	// content is the container that represents the content area
	content *fyne.Container

	// the objects below are all parts of the navbar area
	searchEntry     *widget.Entry
	typeSelectEntry *widget.Select
	addItemButton   fyne.CanvasObject
	itemsWidget     *itemsWidget
}

func newVaultView(mw *mainView, vault *azure.SecretsVault, name string) *vaultView {
	vw := &vaultView{
		mainView: mw,
		name: &widget.Label{
			Text:      name,
			Alignment: fyne.TextAlignCenter,
			TextStyle: fyne.TextStyle{Bold: true},
		},
		filterOptions: &paw.VaultFilterOptions{},
		vault:         vault,
	}
	vw.ExtendBaseWidget(vw)

	vw.searchEntry = vw.makeSearchEntry()
	vw.addItemButton = vw.makeAddItemButton()

	vw.itemsWidget = newItemsWidget(vw.vault, vw.filterOptions)
	vw.itemsWidget.OnSelected = func(meta *paw.Metadata) {
		// get metadata from current vault for selected secret
		item, err := vw.vault.GetItem(meta)
		if err != nil {
			dialog.ShowError(err, vw.mainView)
			return
		}
		vw.setContentItem(NewFyneItem(item), vw.itemView)
	}
	vw.typeSelectEntry = vw.makeTypeSelectEntry()
	vw.content = container.NewMax(vw.defaultContent())

	vw.view = container.NewMax(vw.makeView())
	return vw
}

func (vw *vaultView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(vw.view)
}

// Reload reloads the widget according the specified options
func (vw *vaultView) Reload() {
	vw.typeSelectEntry = vw.makeTypeSelectEntry()
	vw.view.Objects[0] = vw.makeView()
	vw.view.Refresh()
}

// emptyVaultContent returns the content to display when the vault has no items
func (vw *vaultView) emptyVaultContent() fyne.CanvasObject {
	msg := fmt.Sprintf("Vault %q is empty", vw.name)
	text := headingText(msg)
	addItemButton := vw.makeAddItemButton()
	return container.NewCenter(container.NewVBox(text, addItemButton))
}

// defaultContent returns the object to display for default states
func (vw *vaultView) defaultContent() fyne.CanvasObject {
	img := canvas.NewImageFromResource(icon.PawIcon)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(64, 64))
	return container.NewCenter(img)
}

// setContent sets the content view for an Item action (i.e. view or edit) and handle the context view (creation and cancellation)
func (vw *vaultView) setContentItem(item FyneItem, f func(context.Context, FyneItem) fyne.CanvasObject) {
	if vw.cancelCtx != nil {
		vw.cancelCtx()
	}
	ctx, cancel := context.WithCancel(context.Background())
	vw.cancelCtx = cancel
	o := f(ctx, item)
	vw.content.Objects = []fyne.CanvasObject{o}
	vw.content.Refresh()
}

// setContent sets the content view with the provided object and refresh
func (vw *vaultView) setContent(o fyne.CanvasObject) {
	if vw.cancelCtx != nil {
		vw.cancelCtx()
	}
	vw.cancelCtx = nil
	vw.content.Objects = []fyne.CanvasObject{o}
	vw.content.Refresh()
}

// makeView returns the view container
func (vw *vaultView) makeView() fyne.CanvasObject {
	left := container.NewBorder(container.NewVBox(vw.makeVaultMenu(), vw.searchEntry), vw.addItemButton, nil, nil, vw.itemsWidget)
	split := container.NewHSplit(left, vw.content)
	split.Offset = 0.3
	return split
}

func (vw *vaultView) makeVaultMenu() fyne.CanvasObject {
	switchVault := fyne.NewMenuItem("Switch Vault", func() {
		vw.mainView.Reload()
	})

	vaults := vw.mainView.Conf.Vaults
	if len(vaults) == 1 {
		switchVault.Disabled = true
	}

	return container.NewBorder(nil, nil, nil, nil, vw.name)
}

// makeSearchEntry returns the search entry used to filter the item list by name
func (vw *vaultView) makeSearchEntry() *widget.Entry {
	search := widget.NewEntry()
	search.SetPlaceHolder("Search")
	search.SetText(vw.filterOptions.Name)
	search.OnChanged = func(s string) {
		vw.filterOptions.Name = s
		vw.itemsWidget.Reload(nil, vw.filterOptions)
	}
	return search
}

// makeTypeSelectEntry returns the select entry used to filter the item list by type
func (vw *vaultView) makeTypeSelectEntry() *widget.Select {
	itemTypeMap := map[string]paw.ItemType{}
	options := []string{fmt.Sprintf("All items (%d)", vw.vault.Size())}
	for _, item := range vw.makeItems() {
		i := item
		t := i.GetMetadata().Type
		name := fmt.Sprintf("%s (%d)", strings.Title(t.String()), vw.vault.SizeByType(t))
		options = append(options, name)
		itemTypeMap[name] = t
	}

	filter := widget.NewSelect(options, func(s string) {
		var v paw.ItemType
		if s == options[0] {
			v = paw.ItemType(0) // No item type will be selected
		} else {
			v = itemTypeMap[s]
		}

		vw.filterOptions.ItemType = v
		vw.itemsWidget.Reload(nil, vw.filterOptions)
	})

	filter.SetSelectedIndex(0)
	return filter
}

// makeItems returns a slice of empty paw.Item ready to use as template for
// item's creation
func (vw *vaultView) makeItems() []paw.Item {
	note := paw.NewNote()
	password := paw.NewPassword()
	website := paw.NewLogin()

	return []paw.Item{
		note,
		password,
		website,
	}
}

// makeAddItemButton returns the button used to add an item to the vault
func (vw *vaultView) makeAddItemButton() fyne.CanvasObject {

	button := widget.NewButtonWithIcon("New Entry", theme.ContentAddIcon(), func() {
		// on button press we need to display the content of password menu
		p := paw.NewLogin()
		vw.setContentItem(NewFyneItem(p), vw.editItemView)
		vw.Reload()
	})
	button.Importance = widget.HighImportance
	return button
}

// itemView returns the view that displays the item's content along with the allowed actions
func (vw *vaultView) itemView(ctx context.Context, fyneItem FyneItem) fyne.CanvasObject {
	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), func() {
		vw.setContentItem(fyneItem, vw.editItemView)
	})
	var bottom *fyne.Container
	top := container.NewBorder(nil, nil, nil, editBtn, widget.NewLabel(""))

	content := fyneItem.Show(ctx, vw.mainView.Window)
	// progress is global
	meta := ShowMetadata(fyneItem.Item().GetMetadata())
	bottom = container.NewGridWithColumns(2, meta, progress)

	return container.NewBorder(top, bottom, nil, nil, content)
}

// editItemView returns the view that allow to edit an item
func (vw *vaultView) editItemView(ctx context.Context, fyneItem FyneItem) fyne.CanvasObject {

	var isNew bool
	item := fyneItem.Item()
	metadata := item.GetMetadata()

	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		vw.cancelCtx()
		if metadata.Created.IsZero() {
			vw.setContent(vw.defaultContent())
			return
		}
		vw.setContentItem(fyneItem, vw.itemView)
	})

	if metadata.Created.IsZero() {
		isNew = true
	}

	content, editItem := fyneItem.Edit(ctx, vw.vault.Key(), vw.mainView.Window)
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		metadata := editItem.GetMetadata()

		// TODO: update to use the built-in entry validation
		if metadata.Name == "" {
			d := dialog.NewInformation("", "The title cannot be emtpy", vw.mainView.Window)
			d.Show()
			return
		}

		var reloadItems bool

		// add item to vault
		err := vw.vault.AddItem(editItem)
		if err != nil {
			dialog.ShowError(err, vw.mainView)
			return
		} else {
			reloadItems = true
		}

		if metadata.Favicon != editItem.GetMetadata().Favicon {
			reloadItems = true
		}

		item = editItem

		if reloadItems {
			vw.itemsWidget.Reload(item, vw.filterOptions)
		}

		fyneItem := NewFyneItem(item)
		vw.setContentItem(fyneItem, vw.itemView)
		vw.Reload()

	})
	saveBtn.Importance = widget.HighImportance
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		msg := widget.NewLabel(fmt.Sprintf("Are you sure you want to delete %q?", item.String()))
		d := dialog.NewCustomConfirm("", "Delete", "Cancel", msg, func(b bool) {
			if b {
				err := vw.vault.DeleteItem(editItem)
				if err != nil {
					dialog.ShowError(err, vw.mainView)
					return
				}
				vw.itemsWidget.Reload(nil, vw.filterOptions)
				vw.setContent(vw.defaultContent())
				vw.Reload()
			}
		}, vw.mainView.Window)
		d.Show()
	})

	saveCenter := container.NewHBox(saveBtn, deleteBtn)
	if isNew {
		saveCenter = container.NewHBox(saveBtn)
	}
	top := container.NewBorder(nil, nil, cancelBtn, saveCenter, layout.NewSpacer())

	// elements should not be displayed on create but only on edit
	var bottomContent fyne.CanvasObject
	if !metadata.Created.IsZero() {
		bottomContent = ShowMetadata(metadata)
	}

	//bottom := container.NewBorder(bottomContent, nil, nil, nil, layout.NewSpacer())
	return container.NewBorder(top, bottomContent, nil, nil, content)
}
