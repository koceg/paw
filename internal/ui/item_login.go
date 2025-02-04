package ui

import (
	"bytes"
	"context"
	"image/png"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"lucor.dev/paw/internal/favicon"
	"lucor.dev/paw/internal/icon"
	"lucor.dev/paw/internal/paw"
)

// Declare conformity to Item interface
var _ paw.Item = (*Login)(nil)

// Declare conformity to FyneItem interface
var _ FyneItem = (*Login)(nil)

type Login struct {
	*paw.Login
}

func (login *Login) Item() paw.Item {
	return login.Login
}

func (login *Login) Icon() fyne.Resource {
	if login.Favicon != nil {
		return login.Favicon
	}
	return icon.KeyOutlinedIconThemed
}

func (login *Login) Edit(ctx context.Context, key *paw.Key, w fyne.Window) (fyne.CanvasObject, paw.Item) {

	loginIcon := widget.NewIcon(login.Icon())

	loginItem := &paw.Login{}
	*loginItem = *login.Login
	loginItem.Metadata = &paw.Metadata{}
	*loginItem.Metadata = *login.Metadata
	loginItem.Note = &paw.Note{}
	*loginItem.Note = *login.Note
	loginItem.Password = &paw.Password{}
	*loginItem.Password = *login.Password

	passwordBind := binding.BindString(&loginItem.Password.Value)

	titleEntry := widget.NewEntryWithData(binding.BindString(&loginItem.Name))
	titleEntry.Validator = nil
	titleEntry.PlaceHolder = "Untitled login"

	urlEntry := newURLEntryWithData(ctx, binding.BindString(&loginItem.URL))
	urlEntry.TitleEntry = titleEntry
	urlEntry.FaviconListener = func(favicon *paw.Favicon) {
		loginItem.Metadata.Favicon = favicon
		if favicon != nil {
			loginIcon.SetResource(favicon)
			return
		}
		// no favicon found, fallback to default
		loginIcon.SetResource(icon.KeyOutlinedIconThemed)
	}

	usernameEntry := widget.NewEntryWithData(binding.BindString(&loginItem.Username))
	usernameEntry.Validator = nil

	// the note field
	noteEntry := widget.NewEntryWithData(binding.BindString(&loginItem.Note.Value))
	noteEntry.MultiLine = true
	noteEntry.Validator = nil

	// center
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.Bind(passwordBind)
	passwordEntry.Validator = nil
	passwordEntry.SetPlaceHolder("Password")

	passwordCopyButton := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		w.Clipboard().SetContent(passwordEntry.Text)
	})

	passwordMakeButton := widget.NewButtonWithIcon("Generate", icon.KeyOutlinedIconThemed, func() {
		pg := NewPasswordGenerator(key)
		pg.ShowPasswordGenerator(passwordBind, loginItem.Password, w)
	})

	form := container.New(layout.NewFormLayout())
	form.Add(loginIcon)
	form.Add(titleEntry)

	form.Add(labelWithStyle("URL"))
	form.Add(urlEntry)

	form.Add(labelWithStyle("Username"))
	form.Add(usernameEntry)

	form.Add(labelWithStyle("Password"))

	form.Add(container.NewBorder(nil, nil, nil, container.NewHBox(passwordCopyButton, passwordMakeButton), passwordEntry))

	form.Add(labelWithStyle("Note"))
	form.Add(noteEntry)

	return form, loginItem
}

func (login *Login) Show(ctx context.Context, w fyne.Window) fyne.CanvasObject {
	obj := titleRow(login.Icon(), login.Name)
	if login.URL != "" {
		obj = append(obj, copiableLinkRow("URL", login.URL, w)...)
	}
	if login.Username != "" {
		obj = append(obj, copiableRow("Username", login.Username, w)...)
	}
	if login.Password.Value != "" {
		obj = append(obj, copiablePasswordRow("Password", login.Password.Value, w)...)
	}
	if login.Note != nil && login.Note.Value != "" {
		obj = append(obj, copiableRow("Note", login.Note.Value, w)...)
	}
	return container.New(layout.NewFormLayout(), obj...)
}

type urlEntry struct {
	widget.Entry
	TitleEntry      *widget.Entry
	FaviconListener func(*paw.Favicon)

	ctx  context.Context
	host string // host keep track of the initial value before editing
}

func newURLEntryWithData(ctx context.Context, bind binding.String) *urlEntry {
	e := &urlEntry{
		ctx: ctx,
	}
	e.ExtendBaseWidget(e)
	e.Bind(bind)
	e.Validator = nil

	rawurl, _ := bind.Get()

	e.host = e.hostFromRawURL(rawurl)
	return e
}

func (e *urlEntry) hostFromRawURL(rawurl string) string {
	if rawurl == "" {
		return rawurl
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}

	if u.Host != "" {
		return u.Host
	}
	parts := strings.Split(u.Path, "/")
	return parts[0]
}

// FocusLost is a hook called by the focus handling logic after this object lost the focus.
func (e *urlEntry) FocusLost() {
	defer e.Entry.FocusLost()

	host := e.hostFromRawURL(e.Text)
	if e.TitleEntry.Text == "" {
		e.TitleEntry.SetText(host)
	}
	if host == e.host {
		return
	}
	e.host = host

	go func() {
		var fav *paw.Favicon

		img, err := favicon.Download(e.ctx, host, favicon.Options{
			ForceMinSize: true,
		})
		if err != nil {
			e.FaviconListener(fav)
			return
		}

		w := &bytes.Buffer{}
		err = png.Encode(w, img)
		if err != nil {
			e.FaviconListener(fav)
			return
		}

		fav = paw.NewFavicon(host, w.Bytes())
		e.FaviconListener(fav)
	}()
}
