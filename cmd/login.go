package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joshuapmorgan/forgerock-aws-sts/aws"
	"github.com/joshuapmorgan/forgerock-aws-sts/config"
	"github.com/joshuapmorgan/forgerock-aws-sts/idp"
	"github.com/joshuapmorgan/forgerock-aws-sts/saml"
)

type LoginCommand struct {
	Meta
}

func (c *LoginCommand) Run(args []string) int {
	co, err := config.Load()
	if err != nil {
		if os.IsNotExist(err) {
			c.Ui.Error(fmt.Sprintf("forgerock-aws-sts is not yet configured - please run the configure sub-command to configure."))
			return 1
		}
		c.Ui.Error(fmt.Sprintf("Failed to load configuration: %s", err))
		return 1
	}

	var cfg idp.Config
	cfg.BaseUri = co.BaseURI
	cfg.Realm = co.MetaAlias
	cfg.SPEntityId = co.SPEntityID

	sess, err := idp.New(cfg)
	if err := sess.Login(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error while attempting logon initiation: %s", err))
		return 1
	}

	for sess.HasMoreRequirements() {
		var cb []idp.Callback
		for _, req := range sess.GetRequirements() {
			switch req := req.(type) {
			case idp.NameCallback:
				name, err := c.Ui.Ask(req.Prompt())
				if err != nil {
					c.Ui.Error(fmt.Sprintf("Error while asking for variable input: %s", err))
					return 1
				}
				req.SetName(name)
				cb = append(cb, req)
			case idp.PasswordCallback:
				password, err := c.Ui.AskSecret(req.Prompt())
				if err != nil {
					c.Ui.Error(fmt.Sprintf("Error while asking for variable input: %s", err))
					return 1
				}
				req.SetPassword(password)
				cb = append(cb, req)
			case idp.ConfirmationCallback:
				cb = append(cb, req)
			default:
				c.Ui.Error("Unknown callback type encountered.")
				return 1
			}
		}
		if err := sess.SubmitRequirements(cb); err != nil {
			c.Ui.Error(fmt.Sprintf("Error while submitting requirements: %s", err))
			return 1
		}
	}

	if sess.GetStatus() != idp.SUCCESS {
		c.Ui.Error("Authentication Failure")
		return 1
	}

	assertion, err := sess.IDPSSOInit()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error while performing SSO initialisation: %s", err))
		return 1
	}

	response, err := saml.ParseEncodedSAMLResponse(assertion)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error while parsing encoded SAML response: %s", err))
		return 1
	}

	attrs := response.GetAttributeValues("https://aws.amazon.com/SAML/Attributes/Role")

	roles := []*saml.SAMLRole{}
	for _, role := range attrs {
		x := strings.Split(role, ",")
		role_arn := x[0]
		principal_arn := x[1]

		role := &saml.SAMLRole{
			RoleArn:      role_arn,
			PrincipalArn: principal_arn,
		}

		roles = append(roles, role)
	}

	c.Ui.Output("Please choose the role you would like to assume:")

	for i, role := range roles {
		c.Ui.Output(fmt.Sprintf("[%d] %s", i, role.RoleArn))
	}

	s, err := c.Ui.Ask("Selection:")
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error while asking for variable input: %s", err))
		return 1
	}

	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error while converting selected role input to integer: %s", err))
		return 1
	}

	if i > len(roles)-1 || i < 0 {
		c.Ui.Error(fmt.Sprintf("Invalid selection entered: %d", i))
		return 1
	}

	credentials, err := aws.LoginToSTSUsingRole(roles[i], assertion)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error while login to STS using role: %s", err))
		return 1
	}

	c.Ui.Output("")
	c.Ui.Output(fmt.Sprintf("Access Key ID: %s", credentials.AccessKeyId))
	c.Ui.Output(fmt.Sprintf("Secret Access Key: %s", credentials.SecretAccessKey))
	c.Ui.Output(fmt.Sprintf("Session Token: %s", credentials.SessionToken))
	c.Ui.Output("")
	c.Ui.Output(fmt.Sprintf("Expiration at %s", credentials.Expiration.Local().String()))

	return 0
}

func (c *LoginCommand) Help() string {
	helpText := `
Usage: forgerock-aws-sts login

  Log into Forgerock based IdP and use SAML assertion to request temporary AWS API keys using AWS STS
`

	return strings.TrimSpace(helpText)
}

func (c *LoginCommand) Synopsis() string {
	return "Log into Forgerock based IdP and use SAML assertion to request temporary AWS API keys using AWS STS"
}
