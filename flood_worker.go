package main

import (
	"bytes"
	"crypto/tls"
	"net/http"
)

type floodWorker struct {
	dead           bool
	exitChan       chan int
	id             int
	RequestCounter int
}

func (fw *floodWorker) Start() {
	go func() {
		defer fw.Kill()
		client := &http.Client{}
		if scheme == "https" {
			//Skip certificate verify for performance
			secureTransport := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client = &http.Client{Transport: secureTransport}
		}

		for {
			if fw.dead {
				return
			}

			body := []byte(tokenizedBody.String())
			req, _ := http.NewRequest(*method, tokenizedTarget.String(), bytes.NewBuffer(body))
			req.Header.Set("User-Agent", getRandomUserAgent())
			if *method == "POST" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			//Inject custom headers right before sending
			injectHeaders(req)
			resp, err := client.Do(req)
			if err != nil {
				lastErr = err.Error()
			}
			// Close body to prevent a "too many files open" error
			err = resp.Body.Close()
			if err != nil && lastErr == "" {
				lastErr = err.Error()
			}
			fw.RequestCounter += 1 //Worker specific counter
			requestChan <- true
		}
	}()
}

func (fw *floodWorker) Kill() {
	fw.dead = true
	fw.exitChan <- fw.id
}
