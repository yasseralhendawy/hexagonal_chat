package main

import (
	"fmt"

	"github.com/yasseralhendawy/hexagonal_chat/config"
)

func main() {
	//lets get the configrations for now just check the error
	_, err := config.GetConfig("/exp1")
	if err != nil {
		panic(err)
	}
	fmt.Println("Hello world from auth")
}
