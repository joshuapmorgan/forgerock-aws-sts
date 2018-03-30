package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/joshuapmorgan/forgerock-aws-sts/config"
)

type ConfigureCommand struct {
	Meta
}

func (c *ConfigureCommand) Run(args []string) int {
	baseUri, err := c.Ui.Ask("Base URI:")
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error while asking for variable input: %s", err))
		return 1
	}

	uri, err := url.Parse(baseUri)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Invalid base URI: %s", err))
		return 1
	}

	if uri.String() == "" {
		c.Ui.Error("Cannot specify empty base URI.")
		return 1
	}

	metaAlias, err := c.Ui.Ask("Meta Alias/Realm:")
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error while asking for variable input: %s", err))
		return 1
	}

	if metaAlias == "" {
		c.Ui.Error("Cannot specify empty meta alias.")
		return 1
	}

	spEntityId, err := c.Ui.Ask("SP Entity ID:")
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error while asking for variable input: %s", err))
		return 1
	}

	if spEntityId == "" {
		c.Ui.Error("Cannot specify empty SPEntityId.")
		return 1
	}

	cfg := config.Config{
		BaseURI:    uri.String(),
		MetaAlias:  metaAlias,
		SPEntityID: spEntityId,
	}

	if err = config.Save(&cfg); err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to save configuration: %s", err))
		return 1
	}

	c.Ui.Output("Successfully saved configuration.")
	return 0
}

func (c *ConfigureCommand) Help() string {
	helpText := `
Usage: forgerock-aws-sts configure

  Configures forgerock-aws-sts
`

	return strings.TrimSpace(helpText)
}

func (c *ConfigureCommand) Synopsis() string {
	return "Configures forgerock-aws-sts"
}
