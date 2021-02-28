package logging

import (
	"github.com/47-11/spotifete/config"
	"github.com/google/logger"
	"io"
	"io/ioutil"
	"log"
	"sync"
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
		return openLogFile("spotifete.log")
	} else {
		return ioutil.Discard
	}
}
