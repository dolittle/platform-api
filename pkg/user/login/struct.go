package login

type LoginConfiguration struct {
	Serve      Serve      `yaml:"serve"`
	Urls       Urls       `yaml:"urls"`
	Identities Identities `yaml:"identities"`
	Providers  Providers  `yaml:"providers"`
	Flows      Flows      `yaml:"flows"`
	Clients    Clients    `yaml:"clients"`
}
type Serve struct {
	Port    int    `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}
type Urls struct {
	Error string `yaml:"error"`
}

type Tenants map[string]string

type Identities struct {
	CookieName string  `yaml:"cookie_name"`
	Tenants    Tenants `yaml:"tenants"`
}
type Microsoft struct {
	Name     string `yaml:"name"`
	ImageURL string `yaml:"image_url"`
}
type Github struct {
	Name     string `yaml:"name"`
	ImageURL string `yaml:"image_url"`
}
type Providers struct {
	Microsoft Microsoft `yaml:"microsoft"`
	Github    Github    `yaml:"github"`
}
type Login struct {
	FlowIDQueryParameter string `yaml:"flow_id_query_parameter"`
	CsrfTokenParameter   string `yaml:"csrf_token_parameter"`
	ProviderParameter    string `yaml:"provider_parameter"`
}
type Tenant struct {
	FlowIDQueryParameter    string `yaml:"flow_id_query_parameter"`
	FlowIDFormParameter     string `yaml:"flow_id_form_parameter"`
	FlowTenantFormParameter string `yaml:"flow_tenant_form_parameter"`
}
type Consent struct {
	FlowIDQueryParameter string `yaml:"flow_id_query_parameter"`
}
type Flows struct {
	Login   Login   `yaml:"login"`
	Tenant  Tenant  `yaml:"tenant"`
	Consent Consent `yaml:"consent"`
}
type HydraEndpoints struct {
	Admin string `yaml:"admin"`
}
type Hydra struct {
	Endpoints HydraEndpoints `yaml:"endpoints"`
}
type KratosEndpoints struct {
	Public string `yaml:"public"`
}
type Kratos struct {
	Endpoints KratosEndpoints `yaml:"endpoints"`
}
type Clients struct {
	Hydra  Hydra  `yaml:"hydra"`
	Kratos Kratos `yaml:"kratos"`
}
