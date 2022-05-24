package main

import (
	"fmt"

	"github.com/elleven11/gophergotchi/stats"
)

func main() {
	fmt.Printf("%v", stats.GetFeedByUser("cassanof"))
}
