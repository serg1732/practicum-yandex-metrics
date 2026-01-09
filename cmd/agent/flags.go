package main

import "flag"

var remoteAddr string
var reportInterval int
var pollInterval int

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags() {
	flag.StringVar(&remoteAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&reportInterval, "r", 10, "interval between reports")
	flag.IntVar(&pollInterval, "p", 2, "interval between polls")
	flag.Parse()
}
