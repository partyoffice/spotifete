package logging

import (
	"io"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/partyoffice/spotifete/config"
)

var setupGinLogOnce sync.Once

func setupGinLog() {
	setupGinLogOnce.Do(doSetupGinLog)
}

func doSetupGinLog() {
	c := config.Get()
	if c.SpotifeteConfiguration.ReleaseMode {
		gin.DefaultWriter = OpenLogFile("gin/request.log")
		gin.DefaultErrorWriter = io.MultiWriter(OpenLogFile("gin/error.log"), os.Stderr)
	} else {
		gin.DefaultWriter = os.Stdout
		gin.DefaultErrorWriter = os.Stderr
	}
}
