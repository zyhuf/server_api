package logger

import (
	"testing"
	"time"

	"reyzar.com/server-api/pkg/logger"
)

const LOG_TEST_PATH = "/opt/reyzar/nete-agent/testLog/test.log"

func TestLog(t *testing.T) {
	logger.InitLogFile(LOG_TEST_PATH)
	for {
		logger.Debug("sfwerewewrewrewrerewrerwerewreewrsfads234235324324324324fdsfsfe324342343243243dfsf")
		time.Sleep(time.Duration(30) * time.Second)
		//check log
	}
}
