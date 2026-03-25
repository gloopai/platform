package types

type OpsServiceNode struct {
	ServiceName string `json:"service_name"`
	ServiceID   string `json:"service_id"`
	Node        string `json:"node"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Status      string `json:"status"` // passing/warning/critical/unknown
}

type OpsServiceStatus struct {
	ServiceName string           `json:"service_name"`
	Ok          bool             `json:"ok"`
	Total       int              `json:"total"`
	Passing     int              `json:"passing"`
	Warning     int              `json:"warning"`
	Critical    int              `json:"critical"`
	Nodes       []OpsServiceNode `json:"nodes"`
}

type OpsServicesResp struct {
	Ok       bool               `json:"ok"`
	Services []OpsServiceStatus `json:"services"`
}
