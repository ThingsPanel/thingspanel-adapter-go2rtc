package formjson

// SVCRForm 服务接入点凭证表单结构
type SVCRForm struct {
	APIURL       string `json:"api_url"`
	SyncInterval int    `json:"sync_interval"`
	AutoSync     bool   `json:"auto_sync"`
}
