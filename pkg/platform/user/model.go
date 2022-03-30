package user

type KratosUser struct {
	ID        string           `json:"id"`
	SchemaID  string           `json:"schema_id"`
	SchemaURL string           `json:"schema_url"`
	Traits    KratosUserTraits `json:"traits"`
}

type KratosUserTraits struct {
	Email   string   `json:"email"`
	Tenants []string `json:"tenants"`
}
