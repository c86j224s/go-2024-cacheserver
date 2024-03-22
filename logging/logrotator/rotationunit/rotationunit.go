package rotationunit

import "time"

type Unit time.Duration

var (
	Day  = Unit(time.Duration(24) * time.Hour)
	Hour = Unit(time.Hour)
)
