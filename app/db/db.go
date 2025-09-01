package db

import (
	"sync"
	"time"
)

type MapValue struct {
	Value         any
	SetAt         time.Time
	HasExpiryDate bool
	ExpireAt      time.Time
}

type Db struct {
	DbMap map[any]*MapValue

	Mu *sync.Mutex
}

func NewDb() *Db {
	return &Db{
		DbMap: make(map[any]*MapValue),
		Mu:    &sync.Mutex{},
	}

}
