package notify

// Envelope 是 SSE event:data 中的 JSON（与 NSQ 里 envelope 字段一致）。
type Envelope struct {
	Event         string `json:"event"`
	Portal        string `json:"portal"`
	ID            string `json:"id"`
	Title         string `json:"title"`
	Body          string `json:"body"`
	Severity      string `json:"severity"`
	LinkPath      string `json:"link_path"`
	LinkQueryJSON string `json:"link_query_json,omitempty"`
	MetaJSON      string `json:"meta_json,omitempty"`
}
