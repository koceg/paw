package ui

import (
	"log"
	"net/url"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"lucor.dev/paw/internal/azure"
	"lucor.dev/paw/internal/icon"
)

// newVaultView is the function that would get us the visual data of the VAULT
// that we are going to interact withnn

// maxWorkers represents the max number of workers to use in parallel processing
var maxWorkers = runtime.NumCPU()

// progress bar visual representation of Clipboard Timeout
var progress *widget.ProgressBar

// mainView represents the Paw main view
// TODO modify to reflect azure keyvault structure
type mainView struct {
	fyne.Window

	Conf  *azure.Config
	token azcore.TokenCredential

	unlockedVault map[string]*azure.SecretsVault // this act as cache

	view *fyne.Container

	version string
}

// Make returns the fyne user interface
func Make(a fyne.App, w fyne.Window, ver string) fyne.CanvasObject {

	progress = widget.NewProgressBar()
	progress.Hide()
	progress.TextFormatter = func() string {
		return "Clipboard Timeout"
	}

	c, err := azure.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if ver == "" {
		ver = "(unknown)"
	}

	mw := &mainView{
		Window:        w,
		Conf:          c,
		unlockedVault: make(map[string]*azure.SecretsVault),
		version:       ver,
	}

	mw.view = container.NewMax(mw.buildMainView())
	mw.SetMainMenu(mw.makeMainMenu())
	return mw.view
}

func (mw *mainView) setView(v fyne.CanvasObject) {
	mw.view.Objects[0] = v
	mw.view.Refresh()
}

func (mw *mainView) makeMainMenu() *fyne.MainMenu {
	// a Quit item will is appended automatically by Fyne to the first menu item
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Vault", func() {
			mw.setView(mw.createVaultView())
		}),
	)
	switchItem := fyne.NewMenuItem("Switch Vault", func() {
		mw.Reload()
	})

	// config.Vaults can be a function
	if len(mw.Conf.Vaults) <= 1 {
		fileMenu.Items[0].Disabled = true
		switchItem.Disabled = true
	}
	fileMenu.Items = append(fileMenu.Items, switchItem)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			URL := "https://github.com/koceg/pawazure"
			u, _ := url.Parse(URL)
			l := widget.NewLabel("PawAzure - " + mw.version)
			l.Alignment = fyne.TextAlignCenter
			link := widget.NewHyperlink(URL, u)
			link.Alignment = fyne.TextAlignCenter
			co := container.NewCenter(
				container.NewVBox(
					pawLogo(),
					l,
					link,
				),
			)
			d := dialog.NewCustom("About PawAzure", "Ok", co, mw.Window)
			d.Show()
		}),
	)
	return fyne.NewMainMenu(
		fileMenu,

		helpMenu,
	)
}

func (mw *mainView) Reload() {
	mw.setView(mw.buildMainView())
	mw.SetMainMenu(mw.makeMainMenu())
}

func (mw *mainView) buildMainView() fyne.CanvasObject {
	var view fyne.CanvasObject

	vaults := mw.Conf.Vaults
	switch len(vaults) {
	case 0:
		view = mw.initVaultView()
	case 1:
		view = mw.createVaultView()
	default:
		view = mw.vaultListView()
	}
	return view
}

// initVaultView returns the view used to create the first vault
// this is the view if our config file is empty
func (mw *mainView) initVaultView() fyne.CanvasObject {

	logo := pawLogo()

	heading := headingText("Configure Azure Subscription")
	heading.Alignment = fyne.TextAlignCenter

	tenant := widget.NewEntry()
	tenant.SetPlaceHolder("TenantID")

	application := widget.NewEntry()
	application.SetPlaceHolder("ApplicationID")

	// this is the part where we authenticate agains azure
	btn := widget.NewButton("Save", func() {
		mw.Conf.TenantID = tenant.Text
		mw.Conf.ClientID = application.Text
		if err := mw.Conf.WriteConfig(); err != nil {
			dialog.ShowError(err, mw.Window)
			return
		}
		mw.setView(mw.createVaultView())
	})
	btn.Importance = widget.HighImportance

	return container.NewCenter(container.NewVBox(logo, heading, tenant, application, btn))
}

func (mw *mainView) createVaultView() fyne.CanvasObject {
	heading := headingText("Select Azure Key Vault")
	heading.Alignment = fyne.TextAlignCenter

	logo := pawLogo()

	name := widget.NewEntry()
	if len(mw.Conf.Vaults) != 0 {
		name.SetText(mw.Conf.Vaults[0])
	} else {
		name.SetPlaceHolder("Vault Name")
	}
	createButton := widget.NewButtonWithIcon("Open", theme.ContentAddIcon(), func() {
		// TODO: update to use the built-in entry validation
		if name.Text == "" {
			d := dialog.NewInformation("", "The Vault name cannot be emtpy", mw.Window)
			d.Show()
			return
		}
		// we would try and GET the new vault that was selected to be used
		// and if sucessfull save to config file
		if mw.token == nil {
			var err error
			mw.token, err = mw.Conf.NewCredential()
			if err != nil {
				dialog.ShowError(err, mw.Window)
				return
			}
		}
		vault, err := azure.NewSecretsVault(name.Text, mw.token)
		if err != nil {
			dialog.ShowError(err, mw.Window)
			return
		}
		// prevent double write of keyvault
		if mw.Conf.IsAbsent(name.Text) {
			mw.Conf.Vaults = append(mw.Conf.Vaults, name.Text)
			if err := mw.Conf.WriteConfig(); err != nil {
				dialog.ShowError(err, mw.Window)
				return
			}
		}
		mw.unlockedVault[name.Text] = vault
		// we need to pass the name of the vault as well
		mw.setView(newVaultView(mw, vault, name.Text))
		//mw.setView(nil)
		mw.SetMainMenu(mw.makeMainMenu())
	})
	createButton.Importance = widget.HighImportance

	cancelButton := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		mw.Reload()
	})

	return container.NewCenter(container.NewVBox(logo, heading, name, container.NewHBox(cancelButton, createButton)))
}

// vaultListView returns a view with the list of available vaults
func (mw *mainView) vaultListView() fyne.CanvasObject {
	// list the vault that we have saved in the config file ~/.paw/azure.json
	heading := headingText("Azure Key Vaults")
	heading.Alignment = fyne.TextAlignCenter

	logo := pawLogo()

	c := container.NewVBox(logo, heading)

	for _, v := range mw.Conf.Vaults {
		name := v
		resource := icon.LockOpenOutlinedIconThemed
		if _, ok := mw.unlockedVault[name]; !ok {
			resource = icon.LockOutlinedIconThemed
		}
		btn := widget.NewButtonWithIcon(name, resource, func() {
			if mw.token == nil {
				var err error
				mw.token, err = mw.Conf.NewCredential()
				if err != nil {
					dialog.ShowError(err, mw.Window)
					return
				}
			}
			vault, err := azure.NewSecretsVault(name, mw.token)
			if err != nil {
				dialog.ShowError(err, mw.Window)
				return
			}
			mw.unlockedVault[name] = vault
			mw.setView(newVaultView(mw, vault, name))
		})
		btn.Alignment = widget.ButtonAlignLeading
		c.Add(btn)
	}

	return container.NewCenter(c)
}

// vaultView returns the view used to handle a vault
// redundant as once we are authenticated we have access to all known vaults
func (mw *mainView) vaultViewByName(name string) fyne.CanvasObject {
	_, ok := mw.unlockedVault[name]
	if !ok {
		// this needs to be replaced with newVaultView
		// onec we obtain a token
		// return mw.unlockVaultView(name)
		return widget.NewLabel(name)
	}
	//return newVaultView(mw, vault, name)
	return widget.NewLabel(name)
}

func (mw *mainView) LockVault(name string) {
	delete(mw.unlockedVault, name)
	mw.Reload()
}

// headingText returns a text formatted as heading
func headingText(text string) *canvas.Text {
	t := canvas.NewText(text, theme.ForegroundColor())
	t.TextStyle = fyne.TextStyle{Bold: true}
	t.TextSize = theme.TextSubHeadingSize()
	return t
}

// logo returns the Paw logo as a canvas image with the specified dimensions
func pawLogo() *canvas.Image {
	return imageFromResource(icon.PawIcon)
}

func imageFromResource(resource fyne.Resource) *canvas.Image {
	img := canvas.NewImageFromResource(resource)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(64, 64))
	return img
}
