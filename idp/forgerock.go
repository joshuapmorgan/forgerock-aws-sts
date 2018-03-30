package idp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type Callback interface {
	Prompt() string
	CallbackType() CallbackType
}

type NameCallback struct {
	name   string
	value  string
	prompt string
}

type PasswordCallback struct {
	name   string
	value  string
	prompt string
}

type ConfirmationCallback struct {
	name  string
	value string
}

type Status int
type CallbackType int

const (
	SUCCESS Status = iota
	FAILURE
	IN_PROGRESS
)

const (
	NAME_CALLBACK CallbackType = iota
	PASSWORD_CALLBACK
	CONFIRMATION_CALLBACK
)

func (c NameCallback) Prompt() string {
	return c.prompt
}

func (c *NameCallback) SetName(name string) {
	c.value = name
}

func (c *NameCallback) Name() string {
	return c.name
}

func (c *NameCallback) Value() string {
	return c.value
}

func (c NameCallback) CallbackType() CallbackType {
	return NAME_CALLBACK
}

func (c PasswordCallback) Prompt() string {
	return c.prompt
}

func (c *PasswordCallback) SetPassword(password string) {
	c.value = password
}

func (c *PasswordCallback) Name() string {
	return c.name
}

func (c *PasswordCallback) Value() string {
	return c.value
}

func (c PasswordCallback) CallbackType() CallbackType {
	return PASSWORD_CALLBACK
}

func (c ConfirmationCallback) Prompt() string {
	return ""
}

func (c *ConfirmationCallback) Name() string {
	return "IDToken2"
	//	return c.name
}

func (c ConfirmationCallback) CallbackType() CallbackType {
	return CONFIRMATION_CALLBACK
}

// used to parse JSON response

type CallbackOutput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CallbackInput struct {
	Name  string     `json:"name"`
	Value json.Token `json:"value"`
}

type Callbacks struct {
	CallbackType string           `json:"type"`
	Outputs      []CallbackOutput `json:"output"`
	Inputs       []CallbackInput  `json:"input"`
}

type Response struct {
	AuthId    string      `json:"authId"`
	TokenId   string      `json:"tokenId"`
	Callbacks []Callbacks `json:"callbacks"`
}

type Client struct {
	jar            http.CookieJar
	authId         string
	status         Status
	doc            Response
	httpStatusCode int
	config         Config
}

type Config struct {
	BaseUri    string
	Realm      string
	SPEntityId string
}

func (c Config) AuthenticateURI() (*url.URL, error) {
	s := fmt.Sprintf("%s/openam/json%s/authenticate", c.BaseUri, c.Realm)
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (c Config) IDPSSOInitURI() (*url.URL, error) {
	s := fmt.Sprintf("%s/openam/idpssoinit", c.BaseUri)
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("spEntityID", c.SPEntityId)
	q.Set("redirected", "true")
	m := fmt.Sprintf("%s/idp", c.Realm)
	q.Set("metaAlias", m)
	u.RawQuery = q.Encode()

	return u, nil
}

func New(c Config) (*Client, error) {
	jar, err := cookiejar.New(nil)

	if err != nil {
		return nil, err
	}

	return &Client{
		jar:    jar,
		status: IN_PROGRESS,
		config: c,
	}, nil
}

func (c *Client) Login() error {
	cl := &http.Client{
		Jar: c.jar,
	}

	u, err := c.config.AuthenticateURI()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r Response
	json.NewDecoder(resp.Body).Decode(&r)

	c.doc = r
	c.httpStatusCode = resp.StatusCode

	return nil
}

func (c *Client) HasMoreRequirements() bool {
	if c.GetRequirements() != nil {
		return true
	}
	return false
}

// returns an array of Callback objects
func (c *Client) GetRequirements() []Callback {
	if c.status != IN_PROGRESS {
		return nil
	}

	var callbacks []Callback

	for _, callback := range c.doc.Callbacks {
		var prompt string
		var name string
		for _, output := range callback.Outputs {
			if output.Name == "prompt" {
				prompt = output.Value
			}
		}

		// how to handle this better

		for _, input := range callback.Inputs {
			if input.Value == "" || input.Value == "0" {
				name = input.Name
			}
		}

		if callback.CallbackType == "NameCallback" {
			var newCallback NameCallback
			newCallback.prompt = prompt
			newCallback.name = name
			callbacks = append(callbacks, newCallback)
		}
		if callback.CallbackType == "PasswordCallback" {
			var newCallback PasswordCallback
			newCallback.prompt = prompt
			newCallback.name = name
			callbacks = append(callbacks, newCallback)
		}
		if callback.CallbackType == "ConfirmationCallback" {
			var newCallback ConfirmationCallback
			newCallback.name = name
			callbacks = append(callbacks, newCallback)
		}
	}

	return callbacks
}

// submit populated Callback objects to proceed with auth
func (c *Client) SubmitRequirements(callbacks []Callback) error {
	var r Response
	r.AuthId = c.getAuthId()

	for _, callback := range callbacks {
		if callback.CallbackType() == NAME_CALLBACK {
			var cb Callbacks
			var in CallbackInput
			nameCallback := callback.(NameCallback)
			cb.CallbackType = "NameCallback"
			in.Name = nameCallback.Name()
			in.Value = nameCallback.Value()
			cb.Inputs = append(cb.Inputs, in)
			r.Callbacks = append(r.Callbacks, cb)
		}
		if callback.CallbackType() == PASSWORD_CALLBACK {
			var cb Callbacks
			var in CallbackInput
			passwordCallback := callback.(PasswordCallback)
			cb.CallbackType = "PasswordCallback"
			in.Name = passwordCallback.Name()
			in.Value = passwordCallback.Value()
			cb.Inputs = append(cb.Inputs, in)
			r.Callbacks = append(r.Callbacks, cb)
		}
		if callback.CallbackType() == CONFIRMATION_CALLBACK {
			var cb Callbacks
			var in CallbackInput
			confirmationCallback := callback.(ConfirmationCallback)
			cb.CallbackType = "ConfirmationCallback"
			in.Name = confirmationCallback.Name()
			in.Value = 0
			cb.Inputs = append(cb.Inputs, in)
			r.Callbacks = append(r.Callbacks, cb)

		}
	}

	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	z := bytes.NewReader(b)

	u, err := c.config.AuthenticateURI()
	if err != nil {
		return err
	}

	cl := &http.Client{
		Jar: c.jar,
	}
	req, err := http.NewRequest("POST", u.String(), z)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := cl.Do(req)
	if err != nil {
		return err
	}

	var s Response
	json.NewDecoder(resp.Body).Decode(&s)

	c.doc = s
	c.httpStatusCode = resp.StatusCode
	c.checkAndSetLoginStatus()

	return nil
}

func (c *Client) checkAndSetLoginStatus() {
	if c.doc.TokenId != "" {
		c.status = SUCCESS
	}
	if c.httpStatusCode != 200 {
		c.status = FAILURE
	}
}

func (c *Client) IDPSSOInit() (string, error) {
	if c.status != SUCCESS {
		return "", errors.New("user is not authenticated")
	}

	cl := &http.Client{
		Jar: c.jar,
	}

	u, err := c.config.IDPSSOInitURI()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := cl.Do(req)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	e := doc.Find(`html > body > form > input[name="SAMLResponse"]`)
	saml := e.Eq(0).AttrOr("value", "")

	return saml, nil
}

func (c *Client) GetAuthId() string {
	return c.getAuthId()
}

func (c *Client) getAuthId() string {
	if c.authId == "" {
		c.authId = c.doc.AuthId
	}
	return c.authId
}

func (c *Client) GetStatus() Status {
	return c.status
}
