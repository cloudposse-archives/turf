package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	common "github.com/cloudposse/posse-cli/common/error"
)

func getStsClient(sess *session.Session) *sts.STS {
	return sts.New(sess)
}

func getStsClientWithCreds(sess *session.Session, creds *credentials.Credentials) *sts.STS {
	return sts.New(sess, &aws.Config{Credentials: creds})
}

// GetSession return a new AWS Session
func GetSession() *session.Session {
	session := session.Must(session.NewSession())
	return session
}

// GetCreds return credentials that can be used on a session
func GetCreds(sess *session.Session, role string) *credentials.Credentials {
	creds := stscreds.NewCredentials(sess, role)
	return creds
}

// GetAccountID returns the AWS Account ID of the session
func GetAccountID(sess *session.Session, role string) string {
	var client *sts.STS
	if role == "" {
		client = getStsClient(sess)
	} else {
		creds := GetCreds(sess, role)
		client = getStsClientWithCreds(sess, creds)
	}

	input := sts.GetCallerIdentityInput{}
	ident, err := client.GetCallerIdentity(&input)

	common.AssertErrorNil(err)
	return *ident.Account
}
