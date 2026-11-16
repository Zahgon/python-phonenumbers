// Copyright (C) 2013 The Libphonenumber Authors
// Licensed under the Apache License, Version 2.0 (the "License");
package carrier

import (
	"fmt"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
)

func ExampleNameForNumber() {
	ro := "RO"
	n, _ := phonenumbers.Parse("+40721234567", &ro)
	fmt.Println(mustNameForNumber(n, "en", nil, nil))
	fmt.Println(mustNameForNumber(n, "fr", nil, nil))
	// Output:
	// Vodafone
	// Vodafone
}
