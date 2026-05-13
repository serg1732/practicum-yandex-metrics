package lib

import (
	"log"
	"os"
)

func Worker() {
	os.Exit(1)                             // want `запрещенный вызов os.Exit вне функции main`
	panic("ошибка аналайзера")             // want `запрещенный вызов panic вне функции main`
	log.Fatal("ошибка аналайзера")         // want `запрещенный вызов log.Fatal вне функции main`
	log.Fatalf("ошибка аналайзера: %d", 1) // want `запрещенный вызов log.Fatalf вне функции main`
	log.Fatalln("ошибка аналайзера")       // want `запрещенный вызов log.Fatalln вне функции main`
}
