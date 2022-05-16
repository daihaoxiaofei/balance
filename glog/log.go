package glog

import (
	"github.com/daihaoxiaofei/balance/config"
	"io/ioutil"
	"log"
	"os"
)

var debugLog = log.New(ioutil.Discard, ``, log.LstdFlags|log.Lshortfile)

func init() {
	if config.Run.Debug {
		debugLog.SetOutput(os.Stdout)
	}
}

var (
	Debug  = debugLog.Println
	Debugf = debugLog.Printf
)

func OpenDebug() {
	debugLog.SetOutput(os.Stdout)
}
