package helpers

import (
	"errors"
	"regexp"
)

var numericRegex = regexp.MustCompile(`^\d+$`)
var signedNumericRegex = regexp.MustCompile(`^[-+]?\d+$`)
var boolRegex = regexp.MustCompile(`^(true|false)$`)

// floatRegex accepts plain decimal notation only (no hex floats, Inf or NaN).
var floatRegex = regexp.MustCompile(`^[-+]?(\d+\.?\d*|\.\d+)([eE][-+]?\d+)?$`)
var errInvalidTime = errors.New("invalid time")

const (
	defaultMaxLength = 50
)
