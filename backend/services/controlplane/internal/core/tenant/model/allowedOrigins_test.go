package tenantmodel_test

import (
	"net"
	"testing"
	"time"

	tenantmodel "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	"github.com/stretchr/testify/suite"
)

type OriginsTestSuite struct {
	suite.Suite
	validDomains   []tenantmodel.Origins
	validIPs       []tenantmodel.Origins
	validCIDRs     []tenantmodel.Origins
	invalidDomains []tenantmodel.Origins
	invalidIPs     []tenantmodel.Origins
	invalidCIDRs   []tenantmodel.Origins
}

func (s *OriginsTestSuite) SetupTest() {
	s.validDomains = []tenantmodel.Origins{
		"google.com", // Usando um domínio mais estável para teste pontual se houver rede
	}
	s.validIPs = []tenantmodel.Origins{
		"127.0.0.1",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
	}
	s.validCIDRs = []tenantmodel.Origins{
		"192.168.1.0/24",
		"2001:db8::/32",
	}
	s.invalidDomains = []tenantmodel.Origins{
		"lockari",
		"   ",
	}
	s.invalidIPs = []tenantmodel.Origins{
		"256.256.256.256",
	}
	s.invalidCIDRs = []tenantmodel.Origins{
		"192.168.1.1/33",
	}
}

// hasNetwork checks if we can reach a public DNS resolver
func hasNetwork() bool {
	_, err := net.DialTimeout("tcp", "8.8.8.8:53", 2*time.Second)
	return err == nil
}

func (s *OriginsTestSuite) TestValidDomainsIsValid() {
	if !hasNetwork() {
		s.T().Skip("Skipping network-dependent domain tests")
	}
	for _, origin := range s.validDomains {
		err := origin.IsValid()
		s.NoError(err, "Origin %s should be valid", origin)
	}
}

func (s *OriginsTestSuite) TestInvalidDomainsIsValid() {
	for _, origin := range s.invalidDomains {
		err := origin.IsValid()
		s.Error(err, "Origin %s should be invalid", origin)
	}
}

func (s *OriginsTestSuite) TestValidIPsIs() {
	for _, origin := range s.validIPs {
		err := origin.IsValid()
		s.NoError(err, "IP %s should be valid", origin)
	}
}

func (s *OriginsTestSuite) TestInvalidIPsIs() {
	for _, origin := range s.invalidIPs {
		err := origin.IsValid()
		s.Error(err, "IP %s should be invalid", origin)
	}
}

func (s *OriginsTestSuite) TestValidCIDRsIs() {
	for _, origin := range s.validCIDRs {
		err := origin.IsValid()
		s.NoError(err, "CIDR %s should be valid", origin)
	}
}

func (s *OriginsTestSuite) TestInvalidCIDRsIs() {
	for _, origin := range s.invalidCIDRs {
		err := origin.IsValid()
		s.Error(err, "CIDR %s should be invalid", origin)
	}
}

func TestOriginsTestSuite(t *testing.T) {
	suite.Run(t, new(OriginsTestSuite))
}
