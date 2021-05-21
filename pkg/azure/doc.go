package azure

type ListResponse struct {
	AccountName string   `json:"account_name"`
	ShareName   string   `json:"share_name"`
	Prefix      string   `json:"prefix"`
	Files       []string `json:"files"`
}
