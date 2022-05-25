package main

import (
	"log"
	"os"

	"github.com/elleven11/gophergotchi/ctrl"
	"github.com/elleven11/gophergotchi/model"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	rcontext := model.MakeReqContext(os.Getenv("API_KEY"), os.Getenv("USERNAME"))
	ctrl := ctrl.MakeController(rcontext, "cassanof")

	ctrl.Run()
}
