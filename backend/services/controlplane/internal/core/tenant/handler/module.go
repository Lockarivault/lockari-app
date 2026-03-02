package tenanthandler

type TenantHandler interface {

}

type tenant struct {

}

func NewModule() (TenantHandler,error) {
	m := &tenant{}
	return m, nil
}
