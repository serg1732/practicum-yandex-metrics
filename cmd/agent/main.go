package main

import (
	"log"

	"github.com/serg1732/practicum-yandex-metrics/internal/service"
)

func main() {
	agent := service.BuildCollector("http://localhost:8080/update/%s/%s/%v")
	if err := agent.Run(2, 10); err != nil {
		log.Fatal(err)
	}
}
