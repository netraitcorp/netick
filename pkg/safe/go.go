package safe

import (
	"log"
	"runtime/debug"
)

func Go(g func()) {
	GoWithRecover(g, defaultRecoverHandler)
}

func GoWithRecover(g func(), customRecover func(err interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				customRecover(r)
			}
		}()
		g()
	}()
}

func defaultRecoverHandler(r interface{}) {
	log.Printf("Error in Go routine: %s", r)
	log.Printf("Stack: %s", debug.Stack())
}
