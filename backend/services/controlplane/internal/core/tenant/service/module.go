package tenantservice

type TenantService interface {

}

type tenant struct {

}

func NewModule() (TenantService,error) {
	m := &tenant{}
	return m, nil
}
