package saml

import "encoding/base64"
import "encoding/xml"

type SAMLRole struct {
	RoleArn      string
	PrincipalArn string
}

type Response struct {
	XMLName   xml.Name
	Assertion Assertion `xml:"Assertion"`
}

type Assertion struct {
	XMLName            xml.Name
	AttributeStatement AttributeStatement
}

type AttributeStatement struct {
	XMLName    xml.Name
	Attributes []Attribute `xml:"Attribute"`
}

type Attribute struct {
	XMLName         xml.Name
	Name            string           `xml:",attr"`
	AttributeValues []AttributeValue `xml:"AttributeValue"`
}

type AttributeValue struct {
	XMLName xml.Name
	Type    string `xml:"xsi:type,attr"`
	Value   string `xml:",innerxml"`
}

func ParseEncodedSAMLResponse(b64ResponseXML string) (*Response, error) {
	response := Response{}
	bytesXML, err := base64.StdEncoding.DecodeString(b64ResponseXML)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(bytesXML, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (r *Response) GetAttribute(name string) string {
	for _, attr := range r.Assertion.AttributeStatement.Attributes {
		if attr.Name == name {
			return attr.AttributeValues[0].Value
		}
	}

	return ""
}

func (r *Response) GetAttributeValues(name string) []string {
	var values []string
	for _, attr := range r.Assertion.AttributeStatement.Attributes {
		if attr.Name == name {
			for _, v := range attr.AttributeValues {
				values = append(values, v.Value)
			}
		}
	}

	return values
}
