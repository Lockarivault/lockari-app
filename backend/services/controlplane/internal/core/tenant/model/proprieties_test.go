package tenantmodel_test

import (
	"testing"

	tenantmodel "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	"github.com/stretchr/testify/suite"
)

type ProprietiesTestSuite struct {
	suite.Suite
	nullKey     map[string]any
	normalized  map[string]any
	validValues map[string]any
}

func (s *ProprietiesTestSuite) SetupTest() {
	s.nullKey = nil
	s.normalized = map[string]any{
		"  Key1  ": "value1",
		"KEY2":     "value2",
	}
	s.validValues = map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
}

func (s *ProprietiesTestSuite) TestNewProprietiesTypes() {
	s.Run("nil map", func() {
		p := tenantmodel.NewProprietiesTypes(s.nullKey)
		s.NotNil(p)
		s.Empty(p.ProprietiesKeys)
	})

	s.Run("normalization", func() {
		p := tenantmodel.NewProprietiesTypes(s.normalized)
		s.NotNil(p)
		s.Equal(s.validValues, map[string]any(p.ProprietiesKeys))
	})

	s.Run("empty keys are filtered", func() {
		input := map[string]any{
			"":    "empty",
			"  ":  "spaces",
			"key": "val",
		}
		p := tenantmodel.NewProprietiesTypes(input)
		s.Len(p.ProprietiesKeys, 1)
		s.Equal("val", p.ProprietiesKeys["key"])
	})
}

func (s *ProprietiesTestSuite) TestSetPlatformFullyQualifiedDomain() {
	p := tenantmodel.NewProprietiesTypes(nil)

	s.Run("valid domain", func() {
		err := p.SetPlatformFullyQualifiedDomain("mytenant")
		s.NoError(err)
		s.Equal("mytenant.vault.lockari.com", p.PlatformFullyQualifiedDomain)
	})

	s.Run("invalid domains", func() {
		testCases := []struct {
			name   string
			domain string
		}{
			{"empty", ""},
			{"too short", "ab"},
			{"too long", "a" + string(make([]byte, 101))},
			{"only spaces", "   "},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				err := p.SetPlatformFullyQualifiedDomain(tc.domain)
				s.Error(err)
			})
		}
	})
}

func (s *ProprietiesTestSuite) TestSetAllowedOrigins() {
	p := tenantmodel.NewProprietiesTypes(nil)

	s.Run("valid origins (IP and CIDR)", func() {
		origins := []tenantmodel.Origins{
			"127.0.0.1",
			"192.168.1.0/24",
		}
		err := p.SetAllowedOrigins(origins)
		s.NoError(err)
		// Map key is defined in constants as "allowed_origins"
		s.Equal(origins, p.ProprietiesKeys["allowed_origins"])
	})

	s.Run("invalid origin format", func() {
		origins := []tenantmodel.Origins{
			"not-an-origin",
		}
		err := p.SetAllowedOrigins(origins)
		s.Error(err)
	})

	s.Run("nil origins cleans map", func() {
		err := p.SetAllowedOrigins(nil)
		s.NoError(err)
		s.Empty(p.ProprietiesKeys["allowed_origins"])
	})
}

func (s *ProprietiesTestSuite) TestSetFullyQualifiedDomain() {
	p := tenantmodel.NewProprietiesTypes(nil)
	err := p.SetFullyQualifiedDomain("any")
	s.NoError(err) // Current implementation is a stub returning nil
}

func TestProprietiesTestSuite(t *testing.T) {
	suite.Run(t, new(ProprietiesTestSuite))
}
