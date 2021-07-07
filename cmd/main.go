package main

import "sync"

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go StartWebServer()
	wg.Add(1)
	go StartApiServer()
	wg.Wait()
}
