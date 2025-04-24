package handlers

import "regexp"

var numericRegex = regexp.MustCompile(`^[1-9]\d*$`)
var sanitizeRegex = regexp.MustCompile(`[';<>]|--|/\*|\*/`)
var boolRegex = regexp.MustCompile(`^(true|false)$`)

const (
	iSO8601UTC       = "2006-01-02T15:04:05Z"
	iSO8601WithMS    = "2006-01-02T15:04:05.000Z"
	defaultMaxLength = 50
)

type MaxStringLengthOption func() int

type Optional[T comparable] struct {
	value     T
	present   bool
	hasErrors error
}

func (o Optional[T]) IsPresent() bool {
	return o.present
}
func (o Optional[T]) Get() T {
	if !o.present {
		panic("Get called on Optional with no value")
	}
	return o.value
}
func (o Optional[T]) Error() error {
	if o.hasErrors != nil {
		return o.hasErrors
	}
	return nil
}
