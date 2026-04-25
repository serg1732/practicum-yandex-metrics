package models

// AuditEvent данные события об успешной обработке метрики.
type AuditEvent struct {
	// TS - время события.
	TS int64 `json:"ts"`
	// Metrics - имена метрик.
	Metrics []string `json:"metrics"`
	// IPAddress - адрес от кого пришел запрос.
	IPAddress string `json:"ip_address"`
}
