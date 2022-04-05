package paw

import "strings"

// Declare conformity to Item interface
var _ Item = (*Login)(nil)

type Login struct {
	*Password `json:"password,omitempty"`
	*Note     `json:"note,omitempty"`
	*Metadata `json:"metadata,omitempty"`

	Username string `json:"username,omitempty"`
	URL      string `json:"url,omitempty"`
}

func NewLogin() *Login {
	return &Login{
		Metadata: &Metadata{
			Type: LoginItemType,
		},
		Note:     &Note{},
		Password: &Password{},
	}
}

// would split the string to Username Url and Notes
func (s *Login) SetContent(data *string) {
	if data != nil {
		d := strings.Split(*data, "|")
		if len(d) == 3 {
			s.Username = d[0]
			s.URL = d[1]
			s.Note.Value = d[2]
		}
	}
}
