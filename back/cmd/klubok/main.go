package main

import (
	"log"

	_ "github.com/jackc/pgx/v5/stdlib" // back/internal/application/database.go

	"klubok/internal/application"
)

func main() {
	if err := application.Run(); err != nil {
		log.Fatalln(err)
	}
}
