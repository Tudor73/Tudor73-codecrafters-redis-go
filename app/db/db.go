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
	DbMap        map[any]*MapValue
	ListChannels map[string]chan bool
}

func NewDb() *Db {
	return &Db{
		DbMap:        make(map[any]*MapValue),
		ListChannels: make(map[string]chan bool),
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

func (db *Db) SetValue(key string, value any) {
	db.DbMap[key] = &MapValue{Value: value}
}
