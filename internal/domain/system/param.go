package system

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrParamKeyRequired   = errors.New("param key is required")
	ErrParamValueRequired = errors.New("param value is required")
	ErrParamNotFound      = errors.New("param not found")
)

type Param struct {
	ID        uint64    `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewParam(key, value, remark string) (Param, error) {
	p := Param{
		Key:    strings.TrimSpace(key),
		Value:  strings.TrimSpace(value),
		Remark: strings.TrimSpace(remark),
	}
	if p.Key == "" {
		return Param{}, ErrParamKeyRequired
	}
	if p.Value == "" {
		return Param{}, ErrParamValueRequired
	}
	return p, nil
}
