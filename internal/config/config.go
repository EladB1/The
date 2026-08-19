package config

import (
	"bytes"
	"fmt"
)

type Config struct {
	Debug            *bool
	Strict           bool
	SuppressWarnings bool
	LogBuffer        *bytes.Buffer
}

func (conf Config) PrintDebugLogs() {
	if *conf.Debug && conf.LogBuffer.Len() != 0 {
		fmt.Println(conf.LogBuffer.String())
	}
}
