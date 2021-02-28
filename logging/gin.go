package logging

import (
	"github.com/47-11/spotifete/config"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"sync"
)

var setupGinLogOnce sync.Once

func setupGinLog() {
	setupGinLogOnce.Do(doSetupGinLog)
}

func doSetupGinLog() {
	c := config.Get()
	if c.SpotifeteConfiguration.ReleaseMode {
		gin.DefaultWriter = openLogFile("gin/request.log")
		gin.DefaultErrorWriter = io.MultiWriter(openLogFile("gin/error.log"), os.Stderr)
	} else {
		gin.DefaultWriter = os.Stdout
		gin.DefaultErrorWriter = os.Stderr
	}
}
