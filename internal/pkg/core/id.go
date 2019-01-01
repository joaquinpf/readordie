package core

import (
	"github.com/rs/xid"
	"github.com/speps/go-hashids"
)

// NewID Generates a new random 20 digit string id
func NewID() string {
	return xid.New().String()
}

// NewIDFromData Generates a repeatable 20 digit string id based on the input seed
func NewIDFromData(data []int) string {
	hd := hashids.NewData()
	hd.Salt = "readordie"
	hd.MinLength = 20
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode(data)
	return e
}
