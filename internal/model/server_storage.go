package models

type MemStorage struct {
	GaugeMap   map[string]*Metrics
	CounterMap map[string]*Metrics
}
