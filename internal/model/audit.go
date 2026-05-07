package models

// AuditEvent данные события об успешной обработке метрики.
type AuditEvent struct {
	// IPAddress - адрес от кого пришел запрос.
	IPAddress string `json:"ip_address"`
	// Metrics - имена метрик.
	Metrics []string `json:"metrics"`
	// TS - время события.
	TS int64 `json:"ts"`
}
