package main

import (
	"fmt"

	"github.com/Walther-Knight/blogGATOR/internal/config"
)

func main() {
	temp, _ := config.Read()
	temp.SetUser("brent")
	configFinal, _ := config.Read()
	fmt.Printf("Output of config: %v\n", configFinal)
}
