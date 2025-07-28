package config

import "time"

const (
	UseProxy                  = true
	Headless                  = false
	Deep                      = 1
	TimeOutSec                = time.Second * 0
	Threads                   = 1
	AttemptsToGenerateSession = 3
)
