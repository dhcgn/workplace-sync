package main

import (
	"runtime/debug"
)

func readBuildInfo(key string, trim int) string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == key {
				if trim > 0 && len(setting.Value) > trim {
					return setting.Value[:trim]
				}
				return setting.Value
			}
		}
	}
	return ""
}

var buildInfoCommitID = readBuildInfo("vcs.revision", 16)
var buildInfoTime = readBuildInfo("vcs.time", -1)
var buildInfoModified = readBuildInfo("vcs.modified", -1)
