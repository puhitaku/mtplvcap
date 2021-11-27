package mtp

import log2 "github.com/puhitaku/mtplvcap/log"

var log = func() *log2.Children {
	return log2.GetLogger()
}()
