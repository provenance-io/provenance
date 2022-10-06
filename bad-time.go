package main

import "time"

func main() {
	_ = time.Now() // Need to use time.Now() here.
	// just a time.Now() comment.
	another := time.Now()
	_ = another
}
