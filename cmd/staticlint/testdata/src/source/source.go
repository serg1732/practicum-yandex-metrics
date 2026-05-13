package main

import (
	"log"
	stdlog "log"
	"os"
	operating "os"
)

func main() {
	os.Exit(0)
	panic("разрешена")
	log.Fatal("разрешена")
	log.Fatalf("разрешена: %d", 1)
	log.Fatalln("разрешена")

	func() {
		log.Fatal("ошибка аналайзера") // want `запрещенный вызов log.Fatal вне функции main`
	}()

	func() {
		operating.Exit(1) // want `запрещенный вызов os.Exit вне функции main`
	}()

	go func() {
		log.Fatal("ошибка аналайзера") // want `запрещенный вызов log.Fatal вне функции main`
	}()

	defer func() {
		panic("ошибка аналайзера") // want `запрещенный вызов panic вне функции main`
	}()
}

func worker() {
	os.Exit(1)        // want `запрещенный вызов os.Exit вне функции main`
	operating.Exit(2) // want `запрещенный вызов os.Exit вне функции main`

	panic("ошибка аналайзера") // want `запрещенный вызов panic вне функции main`

	log.Fatal("ошибка аналайзера")         // want `запрещенный вызов log.Fatal вне функции main`
	log.Fatalf("ошибка аналайзера: %d", 1) // want `запрещенный вызов log.Fatalf вне функции main`
	log.Fatalln("ошибка аналайзера")       // want `запрещенный вызов log.Fatalln вне функции main`

	stdlog.Fatal("ошибка аналайзера") // want `запрещенный вызов log.Fatal вне функции main`
}

func init() {
	log.Fatal("ошибка аналайзера") // want `запрещенный вызов log.Fatal вне функции main`
}

func withLoggerInstance() {
	logger := log.New(os.Stderr, "", 0)

	logger.Fatal("ошибка аналайзера")         // want `запрещенный вызов log.Fatal вне функции main`
	logger.Fatalf("ошибка аналайзера: %d", 1) // want `запрещенный вызов log.Fatalf вне функции main`
	logger.Fatalln("ошибка аналайзера")       // want `запрещенный вызов log.Fatalln вне функции main`
}

func nestedAnonymous() {
	func() {
		log.Fatal("ошибка аналайзера") // want `запрещенный вызов log.Fatal вне функции main`
	}()

	go func() {
		log.Fatal("ошибка аналайзера") // want `запрещенный вызов log.Fatal вне функции main`
	}()
}

func shadowedNamesAreIgnored() {
	panic := func(string) {}
	panic("this is not builtin panic")

	log := struct {
		Fatal func(string)
	}{
		Fatal: func(string) {},
	}

	log.Fatal("this is not std log.Fatal")
}
