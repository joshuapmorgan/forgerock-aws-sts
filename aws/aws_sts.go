package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/joshuapmorgan/forgerock-aws-sts/saml"
	"time"
)

type Credentials struct {
	AccessKeyId     string
	Expiration      *time.Time
	SecretAccessKey string
	SessionToken    string
}

func LoginToSTSUsingRole(role *saml.SAMLRole, samlAssertion string) (*Credentials, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	svc := sts.New(sess)
	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  aws.String(role.PrincipalArn),
		RoleArn:       aws.String(role.RoleArn),
		SAMLAssertion: aws.String(samlAssertion),
	}

	resp, err := svc.AssumeRoleWithSAML(params)
	if err != nil {
		return nil, err
	}

	c := &Credentials{
		AccessKeyId:     aws.StringValue(resp.Credentials.AccessKeyId),
		Expiration:      resp.Credentials.Expiration,
		SecretAccessKey: aws.StringValue(resp.Credentials.SecretAccessKey),
		SessionToken:    aws.StringValue(resp.Credentials.SessionToken),
	}

	return c, nil
}
