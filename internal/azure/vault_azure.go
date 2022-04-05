package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"lucor.dev/paw/internal/paw"
	"sort"
	"strings"
	"time"
)

// the main structure MainView would hold the TokenCredential
// and a map of vaults map[string]*SecretsVault
// name here is redundant
type SecretsVault struct {
	client *azsecrets.Client
	// vault holds secrets and would be a cache while the program is active
	secrets map[string]*paw.Login
	key     *paw.Key
}

func NewSecretsVault(name string, cred azcore.TokenCredential) (*SecretsVault, error) {
	vaultURI := "https://" + name + ".vault.azure.net/"
	opt := &azsecrets.ClientOptions{
		azcore.ClientOptions{
			Retry: policy.RetryOptions{
				TryTimeout:    15 * time.Second,
				MaxRetryDelay: 5 * time.Second,
			},
			Telemetry: policy.TelemetryOptions{
				Disabled: true,
			},
		},
	}
	client, err := azsecrets.NewClient(vaultURI, cred, opt)
	if err != nil {
		return nil, err
	}
	key, err := paw.MakeOneTimeKey()
	if err != nil {
		return nil, err
	}
	vault := &SecretsVault{
		client:  client,
		secrets: make(map[string]*paw.Login),
		key:     key,
	}
	vault.getItems()
	return vault, nil
}

// Delete Secret From Vault
func (v *SecretsVault) DeleteItem(secret paw.Item) error {
	s := secret.GetMetadata()
	_, err := v.client.BeginDeleteSecret(context.TODO(), s.Name, nil)
	if err != nil {
		return err
	}
	delete(v.secrets, s.Name)
	return nil
}

// Get Secret From Vault
// NOTE: we might need to return interface type to deal with different objects
func (v *SecretsVault) GetItem(secret paw.Item) (*paw.Login, error) {
	m := secret.GetMetadata()
	if s, ok := v.secrets[m.Name]; ok {
		if s.Password.Value != "" {
			return s, nil
		}
	}
	rsp, err := v.client.GetSecret(context.TODO(), m.Name, nil)
	if err != nil {
		return nil, err
	}
	// create or update
	if _, ok := v.secrets[m.Name]; !ok {
		v.secrets[m.Name] = NewAzureSecret(rsp.Secret)
		v.secrets[m.Name].Password.Value = (*rsp.Secret.Value)
	} else {
		v.secrets[m.Name].Password.Value = (*rsp.Secret.Value)
	}
	return v.secrets[m.Name], nil
}

// initialise keyvault objects
// NOTE: this function is to get data about secrets WITHOUT SECRET VALUE
func (v *SecretsVault) getItems() {
	// default 25 items per call
	index := "secrets/"
	pager := v.client.ListSecrets(nil)
	for pager.NextPage(context.TODO()) {
		for _, s := range pager.PageResponse().Secrets {
			v.secrets[secretName(*s.ID, index)] = NewAzureSecret(s)
		}
	}
}

// list secrets from vault
func (v *SecretsVault) ListItems() []string {
	var secrets []string
	for name := range v.secrets {
		secrets = append(secrets, name)
	}
	sort.Strings(secrets)
	return secrets
}

// Save Secret to Vault
func (v *SecretsVault) AddItem(secret paw.Item) error {
	s := secret.(*paw.Login)
	str := s.Username + "|" + s.URL + "|" + s.Note.Value
	if len(str) > 255 {
		return fmt.Errorf("Concatenation \"Username|URL|Note\" can have 255 chars Max")
	}
	optins := &azsecrets.SetSecretOptions{
		ContentType: &str}
	result, err := v.client.SetSecret(context.TODO(), s.Metadata.Name,
		s.Password.Value, optins)
	if err != nil {
		return err
	}
	// update the attributes of the secret
	v.secrets[s.Metadata.Name] = s
	v.secrets[s.Metadata.Name].Metadata.Created = *result.Attributes.Created
	v.secrets[s.Metadata.Name].Metadata.Modified = *result.Attributes.Updated
	return nil
}

func (v *SecretsVault) Size() int {
	return len(v.secrets)
}

func (v *SecretsVault) SizeByType(_ paw.ItemType) int {
	return v.Size()
}

func (v *SecretsVault) FilterItemMetadata(opt *paw.VaultFilterOptions) []*paw.Metadata {
	metadata := []*paw.Metadata{}
	filter := opt.Name
	for _, secret := range v.secrets {
		if filter != "" && !strings.Contains(secret.Name, filter) {
			continue
		}
		metadata = append(metadata, secret.Metadata)
	}
	// if metadata is empty try to get the secret from azure keyvault
	if len(metadata) == 0 {
		m := &paw.Metadata{
			Name: filter,
		}
		secret, err := v.GetItem(m)
		if err == nil {
			metadata = append(metadata, secret.Metadata)
			return metadata
		}
	}
	sort.Sort(paw.ByString(metadata))
	return metadata
}

func (v *SecretsVault) Key() *paw.Key {
	return v.key
}
