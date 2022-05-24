package main

import (
	"fmt"
	"log"
	"os"

	"github.com/elleven11/gophergotchi/stats"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	rcontext := stats.MakeReqContext(os.Getenv("API_KEY"), os.Getenv("USERNAME"))

	// now do something with s3 or whatever
	fmt.Printf("%v", rcontext.GetFeedByUser("cassanof"))
}
