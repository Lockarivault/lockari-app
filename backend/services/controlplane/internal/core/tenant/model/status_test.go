package tenantmodel

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type StatusTestSuite struct {
	suite.Suite
}

func (suite *StatusTestSuite) TestNewStatusPending() {
	status, err := NewStatus("pending")
	suite.NoError(err)
	suite.Equal(StatusPending, status)
}

func (suite *StatusTestSuite) TestNewStatusActive() {
	status, err := NewStatus("active")
	suite.NoError(err)
	suite.Equal(StatusActive, status)
}

func (suite *StatusTestSuite) TestNewStatusInactive() {
	status, err := NewStatus("inactive")
	suite.NoError(err)
	suite.Equal(StatusInactive, status)
}

func (suite *StatusTestSuite) TestNewStatusFailed() {
	status, err := NewStatus("failed")
	suite.NoError(err)
	suite.Equal(StatusFailed, status)
}

func (suite *StatusTestSuite) TestNewStatusInvalid() {
	status, err := NewStatus("invalid")
	suite.Error(err)
	suite.NotEqual(StatusActive, status)
}

func (suite *StatusTestSuite) TestNewStatusSuspended() {
	status, err := NewStatus("suspended")
	suite.NoError(err)
	suite.Equal(StatusSuspended, status)
}

func (suite *StatusTestSuite) TestIsValid() {
	suite.True(StatusPending.IsValid())
	suite.True(StatusActive.IsValid())
	suite.True(StatusInactive.IsValid())
	suite.True(StatusFailed.IsValid())
	suite.True(StatusSuspended.IsValid())
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, &StatusTestSuite{})
}
