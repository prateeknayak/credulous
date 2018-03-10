package handler

import (
	"os"

	"github.com/golang/glog"
)

func LogAndDieOnFatalError(err error) {
	if err != nil {
		glog.Errorf("ERROR: %s\n", err)
		os.Exit(1)
	}
}
