package tenantmodel

const (
	proprietyPasswordPolicyKey            = "password_policy"
	proprietyCertificateNotificationKey   = "certificate_expiration_notification_days"
	proprietyAllowedClientIPsKey          = "allowed_client_ips"
	proprietyCertificateDefaultExpiration = "certificate_default_expiration"
	proprietyVaultSharingKey              = "vault_sharing"
	proprietyFullyQualifiedDomainKey      = "fully_qualified_domain"
	proprietyAllowedOriginsKey            = "allowed_origins"
	proprietyMaxSecretsKey                = "quota_max_secrets"
	proprietyMaxUsersKey                  = "quota_max_users"
	proprietyMaxStorageBytesKey           = "quota_max_storage_bytes"

	defaultPasswordMinLength           = 12
	defaultCertificateNotificationDays = 30
	defaultCertificateExpirationValue  = 90
	defaultCertificateExpirationUnit   = "days"

	supportedCertificateExpirationUnitDays   = "days"
	supportedCertificateExpirationUnitWeeks  = "weeks"
	supportedCertificateExpirationUnitMonths = "months"
	supportedCertificateExpirationUnitYears  = "years"
)
