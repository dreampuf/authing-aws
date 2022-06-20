package saml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
)

type SAMLResponse struct {
	XMLName      xml.Name `xml:"Response"`
	Text         string   `xml:",chardata"`
	ID           string   `xml:"ID,attr"`
	IssueInstant string   `xml:"IssueInstant,attr"`
	Version      string   `xml:"Version,attr"`
	Destination  string   `xml:"Destination,attr"`
	Samlp        string   `xml:"samlp,attr"`
	Issuer       struct {
		Text string `xml:",chardata"`
		Saml string `xml:"saml,attr"`
	} `xml:"Issuer"`
	Status struct {
		Text       string `xml:",chardata"`
		StatusCode struct {
			Text  string `xml:",chardata"`
			Value string `xml:"Value,attr"`
		} `xml:"StatusCode"`
	} `xml:"Status"`
	Assertion struct {
		Text         string `xml:",chardata"`
		ID           string `xml:"ID,attr"`
		Version      string `xml:"Version,attr"`
		IssueInstant string `xml:"IssueInstant,attr"`
		Saml         string `xml:"saml,attr"`
		Issuer       string `xml:"Issuer"`
		Signature    struct {
			Text       string `xml:",chardata"`
			Ds         string `xml:"ds,attr"`
			SignedInfo struct {
				Text                   string `xml:",chardata"`
				CanonicalizationMethod struct {
					Text      string `xml:",chardata"`
					Algorithm string `xml:"Algorithm,attr"`
				} `xml:"CanonicalizationMethod"`
				SignatureMethod struct {
					Text      string `xml:",chardata"`
					Algorithm string `xml:"Algorithm,attr"`
				} `xml:"SignatureMethod"`
				Reference struct {
					Text       string `xml:",chardata"`
					URI        string `xml:"URI,attr"`
					Transforms struct {
						Text      string `xml:",chardata"`
						Transform []struct {
							Text      string `xml:",chardata"`
							Algorithm string `xml:"Algorithm,attr"`
						} `xml:"Transform"`
					} `xml:"Transforms"`
					DigestMethod struct {
						Text      string `xml:",chardata"`
						Algorithm string `xml:"Algorithm,attr"`
					} `xml:"DigestMethod"`
					DigestValue string `xml:"DigestValue"`
				} `xml:"Reference"`
			} `xml:"SignedInfo"`
			SignatureValue string `xml:"SignatureValue"`
			KeyInfo        struct {
				Text     string `xml:",chardata"`
				X509Data struct {
					Text            string `xml:",chardata"`
					X509Certificate string `xml:"X509Certificate"`
				} `xml:"X509Data"`
			} `xml:"KeyInfo"`
		} `xml:"Signature"`
		Subject struct {
			Text   string `xml:",chardata"`
			NameID struct {
				Text   string `xml:",chardata"`
				Format string `xml:"Format,attr"`
			} `xml:"NameID"`
			SubjectConfirmation struct {
				Text                    string `xml:",chardata"`
				Method                  string `xml:"Method,attr"`
				SubjectConfirmationData struct {
					Text         string `xml:",chardata"`
					Recipient    string `xml:"Recipient,attr"`
					NotOnOrAfter string `xml:"NotOnOrAfter,attr"`
				} `xml:"SubjectConfirmationData"`
			} `xml:"SubjectConfirmation"`
		} `xml:"Subject"`
		Conditions struct {
			Text                string `xml:",chardata"`
			NotBefore           string `xml:"NotBefore,attr"`
			NotOnOrAfter        string `xml:"NotOnOrAfter,attr"`
			AudienceRestriction struct {
				Text     string `xml:",chardata"`
				Audience string `xml:"Audience"`
			} `xml:"AudienceRestriction"`
		} `xml:"Conditions"`
		AuthnStatement struct {
			Text                string `xml:",chardata"`
			AuthnInstant        string `xml:"AuthnInstant,attr"`
			SessionNotOnOrAfter string `xml:"SessionNotOnOrAfter,attr"`
			SessionIndex        string `xml:"SessionIndex,attr"`
			AuthnContext        struct {
				Text                 string `xml:",chardata"`
				AuthnContextClassRef string `xml:"AuthnContextClassRef"`
			} `xml:"AuthnContext"`
		} `xml:"AuthnStatement"`
		AttributeStatement struct {
			Text      string `xml:",chardata"`
			Attribute []struct {
				Text           string `xml:",chardata"`
				Name           string `xml:"Name,attr"`
				NameFormat     string `xml:"NameFormat,attr"`
				Xs             string `xml:"xs,attr"`
				AttributeValue struct {
					Text string `xml:",chardata"`
					Xsi  string `xml:"xsi,attr"`
					Type string `xml:"type,attr"`
				} `xml:"AttributeValue"`
			} `xml:"Attribute"`
		} `xml:"AttributeStatement"`
	} `xml:"Assertion"`
}

func DecodeBase64edSAMLResponse(encoded string) (*SAMLResponse, error) {
	samldecoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode samlresponse occurs error: %w", err)
	}
	var samlassertion SAMLResponse
	err = xml.Unmarshal(samldecoded, &samlassertion)
	if err != nil {
		return nil, fmt.Errorf("unmarshal samlresponse occurs error: %w", err)
	}
	return &samlassertion, nil
}

func ParseArn(assertion *SAMLResponse) (roleArn, principalArn string, err error) {
	for _, attribute := range assertion.Assertion.AttributeStatement.Attribute {
		if attribute.Name != "http://schemas.microsoft.com/ws/2008/06/identity/claims/role" && attribute.Name != "https://aws.amazon.com/SAML/Attributes/Role" {
			continue
		}
		attributeValue := attribute.AttributeValue.Text
		s := strings.Split(attributeValue, ",")
		roleArn = s[0]
		principalArn = s[1]
		break
	}
	if roleArn == "" || principalArn == "" {
		return "", "", fmt.Errorf("can't find role")
	}
	return roleArn, principalArn, nil
}
