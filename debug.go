package neat

import (
	"log"
	"os"
)

var _d = os.Getenv("GNEATDEBUG") != ""

func debug(fmt string, args ...interface{}) {
	if _d {
		log.Printf(fmt, args...)
	}
}
