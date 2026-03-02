package auditlog

type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionRead   Action = "read"
	ActionLogin  Action = "login"
	ActionLogout Action = "logout"
	ActionShare  Action = "share"
	ActionRevoke Action = "revoke"
	ActionGrant  Action = "grant"
	ActionDeny   Action = "deny"
	ActionCopy   Action = "copy"
	ActionPaste  Action = "paste"
	ActionExport Action = "export"
	ActionImport Action = "import"
)

type LoginType string

const (
	LoginTypePassword LoginType = "password"
	LoginTypeMFA      LoginType = "mfa"
	LoginTypeSSO      LoginType = "sso"
)

type SystemResource string

const (
	SystemResourceControlPlane SystemResource = "controlplane"
	SystemResourceDataPlane    SystemResource = "dataplane"
	SystemResourceApp          SystemResource = "app"
)

// UserType defines the type of actor performing the action.
type UserType string

const (
	// UserHuman represents a real person performing an action via UI.
	UserHuman UserType = "human"
	// UserApp represents an application or service token performing an action.
	UserApp UserType = "app"
	// UserSystem represents an automated system process or background job.
	UserSystem UserType = "system"
)
