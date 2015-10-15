package coolbee

import (
	"log"
)

// Logger returns a middleware handler that logs the request as it goes in and the response as it goes out.
func Logger() Handler {
	return func(c Context, log *log.Logger) {
		
	}
}
