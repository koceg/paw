package azure

import (
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"lucor.dev/paw/internal/paw"
	"os"
	"path/filepath"
)

type Config struct {
	SubscriptionID string `json:"subscription_id"`
	TenantID       string `json:"tenant_id"`
	// ClientID is the ID of the application users will authenticate to.
	ClientID string   `json:"application_id"`
	Vaults   []string `json:"azure_vaults"`
}

// read azure config file,create file if non existent
func ReadConfig() (*Config, error) {
	// this would return $HOME/.paw
	jsonConfig := new(Config)
	pawPath, err := paw.NewOSStorage()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(pawPath.Root(), "azure.json")
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, _ := file.Stat()
	// if we get error we need a test statement here
	buff := make([]byte, stat.Size())
	file.Read(buff)
	if len(buff) != 0 {
		if err := json.Unmarshal(buff, jsonConfig); err != nil {
			return nil, err
		}
	}
	return jsonConfig, nil
}

// save azure config file
func (c *Config) WriteConfig() error {
	pawPath, err := paw.NewOSStorage()
	if err != nil {
		return err
	}
	path := filepath.Join(pawPath.Root(), "azure.json")
	m, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, m, 0600)
}

// test if the key vault is present in config file
func (c *Config) IsAbsent(name string) bool {
	for _, v := range c.Vaults {
		if v == name {
			return false
		}
	}
	return true
}

// new azure credentials
func (c *Config) NewCredential() (*azidentity.InteractiveBrowserCredential, error) {
	options := &azidentity.InteractiveBrowserCredentialOptions{
		TenantID: c.TenantID,
		ClientID: c.ClientID,
	}
	cred, err := azidentity.NewInteractiveBrowserCredential(options)
	if err != nil {
		return nil, err
	}
	return cred, err
}
