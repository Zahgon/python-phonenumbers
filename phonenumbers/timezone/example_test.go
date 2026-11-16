// Copyright (C) 2013 The Libphonenumber Authors
// Licensed under the Apache License, Version 2.0 (the "License");
package timezone

import (
	"fmt"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
)

func ExampleTimeZonesForNumber() {
	ro := "RO"
	n, _ := phonenumbers.Parse("+40721234567", &ro)
	zones := mustTimeZonesForNumber(n)
	fmt.Println(len(zones))
	fmt.Println(zones[0])
	us := "US"
	n, _ = phonenumbers.Parse("+976136234567", &us)
	zones = mustTimeZonesForNumber(n)
	fmt.Println(len(zones))
	fmt.Println(zones[0])
	fmt.Println(zones[1])
	// Output:
	// 1
	// Europe/Bucharest
	// 2
	// Asia/Choibalsan
	// Asia/Ulaanbaatar
}
