package model

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"text/template"
)

// Descriptor is the root data type for a HipChat Connect descriptor file.
type Descriptor struct {
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Key          string                  `json:"key"`
	Links        *DescriptorLinks        `json:"links"`
	Vendor       *DescriptorVendor       `json:"vendor"`
	Capabilities *DescriptorCapabilities `json:"capabilities"`
}

// DescriptorVendor describes the vendor offering a given descriptor.
type DescriptorVendor struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// DescriptorLinks contains hypermedia links for the descriptor resource.
type DescriptorLinks struct {
	Self      string `json:"self"`
	API       string `json:"api"`
	Subdomain string `json:"subdomain"`
	Homepage  string `json:"homepage"`
}

// DescriptorCapabilities contains the capabilities section of a descriptor.
type DescriptorCapabilities struct {
	HipChatAPIProvider *DescriptorHipChatAPIProvider `json:"hipchatApiProvider,omitempty"`
	HipChatAPIConsumer *DescriptorHipChatAPIConsumer `json:"hipchatApiConsumer"`
	OAuth2Provider     *DescriptorOAuth2Provider     `json:"oauth2Provider,omitempty"`
	Installable        *DescriptorInstallable        `json:"installable"`
	Configurable       *DescriptorConfigurable       `json:"configurable,omitempty"`
	Webhooks           []*DescriptorWebhook          `json:"webhook,omitempty"`
	Glances            []*DescriptorGlance           `json:"glance,omitempty"`
	WebPanels          []*DescriptorWebPanel         `json:"webPanel,omitempty"`
	Dialogs            []*DescriptorDialog           `json:"dialog,omitempty"`
}

// DescriptorHipChatAPIProvider TODO
type DescriptorHipChatAPIProvider struct {
	URL             string                               `json:"url"`
	AvailableScopes map[string]*DescriptorAvailableScope `json:"availableScopes"`
}

// DescriptorHipChatAPIConsumer TODO
type DescriptorHipChatAPIConsumer struct {
	Avatar string   `json:"avatar"`
	Scopes []string `json:"scopes"`
}

// DescriptorAvailableScope TODO
type DescriptorAvailableScope struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DescriptorOAuth2Provider TODO
type DescriptorOAuth2Provider struct {
	AuthorizationURL string `json:"authorizationUrl"`
	TokenURL         string `json:"tokenUrl"`
}

// DescriptorInstallable TODO
type DescriptorInstallable struct {
	AllowGlobal bool   `json:"allowGlobal"`
	AllowRoom   bool   `json:"allowRoom"`
	CallbackURL string `json:"callbackUrl"`
}

// DescriptorConfigurable TODO
type DescriptorConfigurable struct {
	URL string `json:"url"`
}

// DescriptorWebhook TODO
type DescriptorWebhook struct {
	Name    string `json:"name"`
	Event   string `json:"event"`
	Pattern string `json:"pattern"`
	URL     string `json:"url"`
}

// DescriptorGlance TODO
type DescriptorGlance struct {
	Key      string                 `json:"key"`
	Name     *DescriptorValueHolder `json:"name"`
	QueryURL string                 `json:"queryUrl"`
	Target   string                 `json:"target"`
	Icon     *DescriptorIcon        `json:"icon"`
}

// DescriptorValueHolder TODO
type DescriptorValueHolder struct {
	Value string `json:"value"`
}

// DescriptorIcon TODO
type DescriptorIcon struct {
	URL   string `json:"url"`
	URL2x string `json:"url@2x"`
}

// DescriptorWebPanel TODO
type DescriptorWebPanel struct {
	Key      string                 `json:"key"`
	Name     *DescriptorValueHolder `json:"name"`
	Location string                 `json:"location"`
	URL      string                 `json:"url"`
}

// DescriptorDialog TODO
type DescriptorDialog struct {
	Key     string                     `json:"key"`
	Title   *DescriptorValueHolder     `json:"title"`
	URL     string                     `json:"url"`
	Options *DescriptorWebPanelOptions `json:"options"`
}

// DescriptorWebPanelOptions TODO
type DescriptorWebPanelOptions struct {
	Style            string                               `json:"style"`
	Size             *DescriptorWebPanelSize              `json:"size"`
	SecondaryActions []*DescriptorWebPanelSecondaryAction `json:"secondaryActions"`
}

// DescriptorWebPanelSize TODO
type DescriptorWebPanelSize struct {
	Width  string `json:"width"`
	Height string `json:"height"`
}

// DescriptorWebPanelSecondaryAction TODO
type DescriptorWebPanelSecondaryAction struct {
	Key  string                 `json:"key"`
	Name *DescriptorValueHolder `json:"name"`
}

// DecodeDescriptor decodes a Descriptor instance from a JSON reader
func DecodeDescriptor(r io.Reader) (*Descriptor, error) {
	var d Descriptor
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&d)
	return &d, err
}

// LoadDescriptor loads a descriptor from a file and injects the base URL before
// decoding it from JSON.
func LoadDescriptor(path, baseURL string) (*Descriptor, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	t, err := template.New("descriptor").Parse(string(data))
	if err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	type Context struct{ BaseURL string }
	t.Execute(&b, Context{baseURL})
	return DecodeDescriptor(&b)
}

// MustLoadDescriptor calls LoadDescriptor and panics on an error.
func MustLoadDescriptor(path, baseURL string) *Descriptor {
	d, err := LoadDescriptor(path, baseURL)
	if err != nil {
		panic(err)
	}
	return d
}

// Encode serializes a Descriptor instance to JSON.
func (d *Descriptor) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(d)
}
