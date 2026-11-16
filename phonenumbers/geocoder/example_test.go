// Copyright (C) 2009-2011 The Libphonenumber Authors
// Licensed under the Apache License, Version 2.0 (the "License");
package geocoder

import (
	"fmt"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
)

func ExampleDescriptionForNumber() {
	gb := "GB"
	de := "DE"
	ch := "CH"
	gbn, _ := phonenumbers.Parse("+442083612345", &gb)
	den, _ := phonenumbers.Parse("0891234567", &de)
	chn, _ := phonenumbers.Parse("0431234567", &ch)
	fmt.Println(mustDescriptionForNumber(gbn, "en", nil, nil))
	fmt.Println(mustDescriptionForNumber(gbn, "fr", nil, nil))
	fmt.Println(mustDescriptionForNumber(gbn, "en", nil, &gb))
	us := "US"
	fmt.Println(mustDescriptionForNumber(gbn, "en", nil, &us))
	fmt.Println(mustDescriptionForNumber(den, "en", nil, nil))
	fmt.Println(mustDescriptionForNumber(den, "de", nil, nil) == "München")
	fmt.Println(mustDescriptionForNumber(chn, "de", nil, nil) == "Zürich")
	fmt.Println(mustDescriptionForNumber(chn, "en", nil, nil))
	fmt.Println(mustDescriptionForNumber(chn, "fr", nil, nil))
	fmt.Println(mustDescriptionForNumber(chn, "it", nil, nil))
	// Output:
	// London
	// London
	// London
	// United Kingdom
	// Munich
	// true
	// true
	// Zurich
	// Zurich
	// Zurigo
}
