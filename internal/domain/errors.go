package domain

import "errors"

var (
	ErrInvalidProductID = errors.New("invalid product id")
	ErrDownstream       = errors.New("downstream service error")
)
