package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

var asNameCache map[int]string = map[int]string{}
var netWhoisCache map[string]string = map[string]string{}

type RipeResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	// "messages": [],
	// "see_also": [],
	// "version": "4.1",
	// "data_call_name": "whois",
	// "data_call_status": "supported",
	// "cached": false,
	// "query_id": "20231024103719-c5365a21-5870-4bbc-9a7e-cfeb6301b269",
	// "process_time": 33,
	// "server_id": "app111",
	// "build_version": "live.2023.10.23.178",
	// "time": "2023-10-24T10:37:19.123369"
}

type AsNamesResponse struct {
	RipeResponse
	Data AsNamesResponseData `json:"data"`
}

type AsNamesResponseData struct {
	Names     map[string]string `json:"names"`
	Resources []string          `json:"resources"`
}

type WhoisResponseData struct {
	Records     []WhoisRecordSet `json:"records"`
	IrrRecords  []WhoisRecordSet `json:"irr_records"`
	Authorities []string         `json:"authorities"`
	Resource    string           `json:"resource"`
	// "query_time": "2023-10-24T11:00:00"
}

type WhoisRecordSet []WhoisRecord

func (r WhoisRecordSet) getValue(key string) (string, bool) {
	for _, record := range r {
		if record.Key == key {
			return record.Value, true
		}
	}
	return "", false
}

type WhoisRecord struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	DetailsLink string `json:"details_link"`
}

type WhoisResponse struct {
	RipeResponse
	Data WhoisResponseData `json:"data"`
}

func GetNameForAs(num int) (string, error) {

	name, ok := asNameCache[num]
	if ok {
		return name, nil
	}

	req, err := http.Get(fmt.Sprintf("https://stat.ripe.net/data/as-names/data.json?resource=%d", num))
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	response := AsNamesResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	asNameCache[num] = response.Data.Names[strconv.Itoa(num)]
	return asNameCache[num], nil
}

func GetNameForNet(net string) (string, error) {
	name, ok := netWhoisCache[net]
	if ok {
		return name, nil
	}

	req, err := http.Get(fmt.Sprintf("https://stat.ripe.net/data/whois/data.json?resource=%s", net))
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	response := WhoisResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err

	}

	if len(response.Data.Records) == 0 {
		// No records :(
		netWhoisCache[net] = ""
	} else {
		netname, _ := response.Data.Records[0].getValue("netname")
		descr, ok := response.Data.Records[0].getValue("descr")
		if ok {
			netname += " (" + descr + ")"
		}

		netWhoisCache[net] = netname
	}
	return netWhoisCache[net], nil
}
