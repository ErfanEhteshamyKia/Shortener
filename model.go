package main

import "time"

type URL struct {
	Long      string
	Short     string
	Password  string
	ExpiredAt time.Time `json:"expired_at" bson:"expired_at"`
}
