package tenanttools

type SystemLogEvent struct {
	Module    string
	Service   SystemLogService
	Component SystemLogComponent
	Action    SystemLogAction
	Metadata  map[string]any
}

type SystemLogService map[int]string

func setSystemLogService() SystemLogService {
	return SystemLogService{
		1: "controlplane",
		2: "vault",
		3: "auth",
		4: "billing",
		5: "notification",
		6: "storage",
		7: "sync",
		8: "user",
		9: "webhook",
	}
}

type SystemLogComponent map[int]string

func setSystemLogComponent() SystemLogComponent {
	return SystemLogComponent{
		1: "api",
		2: "cli",
		3: "sdk",
		4: "worker",
	}
}

type SystemLogAction map[int]string

func setSystemLogAction() SystemLogAction {
	return SystemLogAction{
		1: "create",
		2: "update",
		3: "delete",
		4: "read",
	}
}

type SystemLogActorType map[int]string

func setSystemLogActorType() SystemLogActorType {
	return SystemLogActorType{
		1: "user",
		2: "system",
	}
}

func GetSystemLogService(key int) string {
	return setSystemLogService()[key]
}

func GetSystemLogComponent(key int) string {
	return setSystemLogComponent()[key]
}

func GetSystemLogAction(key int) string {
	return setSystemLogAction()[key]
}

func GetSystemLogActorType(key int) string {
	return setSystemLogActorType()[key]
}

func NewSystemLogEvent(module string, service SystemLogService, component SystemLogComponent, action SystemLogAction, metadata map[string]any) *SystemLogEvent {
	return &SystemLogEvent{
		Module:    module,
		Service:   service,
		Component: component,
		Action:    action,
		Metadata:  metadata,
	}
}
