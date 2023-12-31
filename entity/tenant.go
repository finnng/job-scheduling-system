package entity

type Tenant struct {
    Id   int        `json:"id"`
    Type TenantType `json:"type"`
}

type TenantType string

const (
    TenantTypeNew        TenantType = "new"
    TenantTypeSme        TenantType = "sme"
    TenantTypeEnterprise TenantType = "enterprise"
)
