package common

import (
	"os"

	"github.com/sirupsen/logrus"
)

// AssertErrorNil asserts that the error is nil and, if not, exits with an error
func AssertErrorNil(err error) {
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
