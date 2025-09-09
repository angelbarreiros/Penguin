package helpers

import "regexp"

var numericRegex = regexp.MustCompile(`^[1-9]\d*$`)
var sanitizeRegex = regexp.MustCompile(`[';<>]|--|/\*|\*/`)
var boolRegex = regexp.MustCompile(`^(true|false)$`)

const (
	iSO8601UTC       = "2006-01-02T15:04:05Z"
	iSO8601WithMS    = "2006-01-02T15:04:05.000Z"
	defaultMaxLength = 50
)
