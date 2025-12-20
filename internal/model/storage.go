package models

type MemStorage struct {
	GaugeMap   map[string]Gauge
	CounterMap map[string]Counter
}
