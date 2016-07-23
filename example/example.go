package main

import (
	"flag"
	"fmt"

	"github.com/rendom/pokemongo"
)

func main() {
	username := flag.String("u", "", "PTC username")
	password := flag.String("p", "", "PTC username")
	flag.Parse()

	if flag.NFlag() != 2 {
		flag.Usage()
		return
	}

	pgo := pokemongo.New()
	err := pgo.Authenticate(*username, *password)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Auth succeeded! Your token is: %s\n", pgo.GetToken())
}
