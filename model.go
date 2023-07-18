package main

import "time"

type URL struct {
	Long      string
	Short     string
	Password  string
	ExpiredAt time.Time `json:"expired_at" bson:"expired_at"`
}

type Click struct {
	Time  time.Time
	Short string
}

type AnalyticsRequest struct {
	Start  time.Time
	Finish time.Time
}

type AnalyticsResponse struct {
	Clicks int
}
