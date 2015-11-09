package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	Headers map[string]*tokenizedString
)

func loadHeaders() {
	Headers = map[string]*tokenizedString{}

	file, err := os.Open(*headersFile)
	if err != nil {
		//Probably file not found
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			arr := strings.Split(scanner.Text(), ": ")
			Headers[arr[0]] = NewTokenizedString(arr[1])
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
}

func injectHeaders(req *http.Request) {
	for k := range Headers {
		req.Header.Set(k, Headers[k].String())
	}
}
