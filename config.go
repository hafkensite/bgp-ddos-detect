package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var tier1Networks = map[int]string{
	174:  "Cogent",
	646:  "Zayo",
	1299: "Arelion",
	2914: "NTT",
	3257: "GTT",
	3356: "Level 3",
	6830: "Liberty Global",
	6939: "Huricane",
}
var transitNetworks = map[int]string{
	13335:  "Cloudflare, Inc.",
	32787:  "Akamai", // Prolexic
	43366:  "OSSO B.V.",
	200020: "Stichting NBIP-NaWas",
}

var denylistNetworks = map[int]string{
	40410:  "FRIT",                  // akamai noise
	206289: "TSB BANKING GROUP PLC", // akamai noise
}

func configAPI(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(tier1Networks)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	fmt.Fprintf(w, "const tier1Networks = %s;\n", string(data))

	data, err = json.Marshal(transitNetworks)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	fmt.Fprintf(w, "const transitNetworks = %s;\n", string(data))

}
