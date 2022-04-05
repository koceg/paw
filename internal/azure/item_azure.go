package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"lucor.dev/paw/internal/paw"
	"strings"
)

//https://stackoverflow.com/a/46208325
func NewAzureSecret(i interface{}) *paw.Login {
	s := paw.NewLogin()
	index := "secrets/"
	switch v := i.(type) {
	case azsecrets.Item:
		s.SetContent(v.ContentType)
		s.Metadata.Created = *v.Attributes.Created
		s.Metadata.Modified = *v.Attributes.Updated
		s.Metadata.Name = secretName(*v.ID, index)
	case azsecrets.Secret:
		s.SetContent(v.ContentType)
		s.Metadata.Created = *v.Attributes.Created
		s.Metadata.Modified = *v.Attributes.Updated
		s.Metadata.Name = strings.Split(secretName(*v.ID, index), "/")[0]
	}
	return s
}

func secretName(id, index string) string {
	i := strings.LastIndex(id, index)
	return id[i+len(index):]
}
