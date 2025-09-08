package db

import (
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
}

func NewDb() *Db {
	return &Db{
		DbMap: make(map[any]*MapValue),
	}
}

func (db *Db) GetValue(key string) (any, bool) {
	val, ok := db.DbMap[key]
	if !ok {
		return nil, false
	}

	if val.HasExpiryDate && time.Now().After(val.ExpireAt) {
		delete(db.DbMap, key)
		return nil, false
	}
	return val.Value, true
}
