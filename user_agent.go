package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	userAgents []string
	random     *rand.Rand
	source     rand.Source
)

func loadUserAgents() {
	//Load user agents from file
	file, err := os.Open(*userAgentFile)
	if err != nil {
		//File not found, or whatever, use default UA
		userAgents = append(userAgents, "Tsunami Flooder (https://github.com/ammar/tsunami)")
		fmt.Println(err)
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			userAgents = append(userAgents, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	//Initiate random number generator
	source = rand.NewSource(time.Now().UnixNano())
	random = rand.New(source)
}

func getRandomUserAgent() string {
	index := int(random.Uint32()) % len(userAgents)
	return userAgents[index]
}
