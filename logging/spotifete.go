package logging

import (
	"io"
	"io/ioutil"
	"log"
	"sync"

	"github.com/google/logger"
	"github.com/partyoffice/spotifete/config"
)

var setupSpotifeteLogOnce sync.Once

func setupSpotifeteLog() {
	setupSpotifeteLogOnce.Do(doSetupSpotifeteLog)
}

func doSetupSpotifeteLog() {
	logFileWriter := getLogFileWriter()
	logger.Init("spotifete", true, false, logFileWriter)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix | log.Llongfile)
}

func getLogFileWriter() io.Writer {
	c := config.Get()
	if c.SpotifeteConfiguration.ReleaseMode {
		return OpenLogFile("spotifete.log")
	} else {
		return ioutil.Discard
	}
}
