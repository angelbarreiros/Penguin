package helpers

import (
	"errors"
	"regexp"
)

var numericRegex = regexp.MustCompile(`^0|[1-9]\d*$`)
var sanitizeRegex = regexp.MustCompile(`[';<>]|--|/\*|\*/`)
var boolRegex = regexp.MustCompile(`^(true|false)$`)
var floatRegex = regexp.MustCompile(`^[0-9]*\.?[0-9]+$`)
var errInvalidTime = errors.New("invalid time")

const (
	defaultMaxLength = 50
)
