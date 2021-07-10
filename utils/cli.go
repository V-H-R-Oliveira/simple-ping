package utils

import (
	"flag"
	"log"
)

type UserInput struct {
	Domain   string
	Quantity int
}

func NewUserInput(domain string, quantity int) UserInput {
	return UserInput{
		Domain:   domain,
		Quantity: quantity,
	}
}

func GetUserInput() UserInput {
	domain := flag.String("domain", "", "--domain=<ipv4 address|domain name address>")
	quantity := flag.Int("requests", 2, "--requests=<amount of requests>. Default is 2 requests")

	flag.Parse()

	if flag.Parsed() {
		if *domain == "" || *quantity <= 0 {
			flag.PrintDefaults()
			log.Fatal("Invalid options")
		}

		return NewUserInput(*domain, *quantity)
	}

	flag.PrintDefaults()
	return NewUserInput("localhost", *quantity)
}
