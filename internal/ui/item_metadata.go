package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"lucor.dev/paw/internal/icon"
	"lucor.dev/paw/internal/paw"
)

// Declare conformity to Item interface
var _ paw.Item = (*Metadata)(nil)

// Item represents the basic paw identity
type Metadata struct {
	*paw.Metadata
}

func (m *Metadata) Item() paw.Item {
	return m.Metadata
}

func (m *Metadata) Icon() fyne.Resource {
	if m.Favicon != nil {
		return m.Favicon
	}
	if m.Type == paw.LoginItemType {
		return icon.KeyOutlinedIconThemed
	}
	return icon.PawIcon
}

func ShowMetadata(m *paw.Metadata) fyne.CanvasObject {
	ctime := &widget.TextSegment{
		Style: widget.RichTextStyleStrong,
		Text:  "Created: ",
	}
	c_time := &widget.TextSegment{
		Style: widget.RichTextStyleInline,
		Text:  m.Created.Format(time.RFC1123),
	}
	mtime := &widget.TextSegment{
		Style: widget.RichTextStyleStrong,
		Text:  "Modified: ",
	}
	m_time := &widget.TextSegment{
		Style: widget.RichTextStyleInline,
		Text:  m.Modified.Format(time.RFC1123),
	}
	nl := &widget.TextSegment{
		Style: widget.RichTextStyleParagraph,
	}
	return widget.NewRichText(ctime, c_time, nl, mtime, m_time, nl)
}
