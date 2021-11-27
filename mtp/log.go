package mtp

import (
	"github.com/puhitaku/mtplvcap/logging"
)

var log = func() *logging.Children {
	return logging.GetLogger()
}()
