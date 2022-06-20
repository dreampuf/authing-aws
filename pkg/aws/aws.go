package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func FetchCredentialViaSAML(ctx context.Context, region, roleArn, principalArn, assertion string, duration_in_seconds int64) (*sts.Credentials, error) {

	sess := session.Must(session.NewSession(
		defaults.Config(),
		&aws.Config{
			Region: aws.String(region),
		}))
	svc := sts.New(sess)

	input := &sts.AssumeRoleWithSAMLInput{
		DurationSeconds: aws.Int64(duration_in_seconds),
		RoleArn:         aws.String(roleArn),
		PrincipalArn:    aws.String(principalArn),
		SAMLAssertion:   aws.String(assertion),
	}
	res, err := svc.AssumeRoleWithSAMLWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("Failed to AssumeRoleWithSAMLWithContext: %w", err)
	}
	return res.Credentials, nil
}

func DumpCred(cred *sts.Credentials) {
	fmt.Printf("export AWS_ACCESS_KEY_ID=%s\nexport AWS_SECRET_ACCESS_KEY=%s\nexport AWS_SESSION_TOKEN=%s",
		*cred.AccessKeyId, *cred.SecretAccessKey, *cred.SessionToken)
}
