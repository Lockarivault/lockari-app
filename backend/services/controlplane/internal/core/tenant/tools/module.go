package tenanttools

type TenantTools interface {

}

type tenant struct {

}

func NewModule() (TenantTools,error) {
	m := &tenant{}
	return m, nil
}
