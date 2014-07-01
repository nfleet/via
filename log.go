package main

import "log"

type Debugging bool

func (d Debugging) Printf(format string, args ...interface{}) {
	if d {
		log.Printf(format, args...)
	}
}

func (d Debugging) Println(args ...interface{}) {
	if d {
		log.Println(args...)
	}
}
