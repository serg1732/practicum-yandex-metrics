package models

type MemStorage struct {
	GaugeMap   map[string]*float64
	CounterMap map[string]*int64
}
