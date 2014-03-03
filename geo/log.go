package geo

import "log"

func (d debugging) Printf(format string, args ...interface{}) {
	if d {
		log.Printf(format, args...)
	}
}

func (d debugging) Println(args ...interface{}) {
	if d {
		log.Println(args...)
	}
}
