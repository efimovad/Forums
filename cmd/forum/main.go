package main

import (
	"github.com/efimovad/Forums.git/internal/app"
	"log"
)

func main() {
	if err := app.Start(); err != nil {
		log.Println(err)
	}
}
