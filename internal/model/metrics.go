package models

// Типы метрик.
const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	// ID - имя метрики.
	ID string `json:"id"`
	// MType - тип метрики.
	MType string `json:"type"`
	// Delta - значение для counter метрики.
	Delta *int64 `json:"delta,omitempty"`
	// Value - значение для gauge метрики.
	Value *float64 `json:"value,omitempty"`
	// Hash - hash значение.
	Hash string `json:"hash,omitempty"`
}
