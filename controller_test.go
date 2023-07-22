package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMain(m *testing.M) {
	cancel, disconnect := setupDB(true)
	defer cancel()
	defer disconnect()

	code := m.Run()

	collection.DeleteMany(context.TODO(), bson.D{})

	os.Exit(code)
}

func TestShortenRouteSuccess(t *testing.T) {
	r := setupRouter()

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(URL{Long: "https://google.com", Short: "abc"})
	if err != nil {
		panic(err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/shorten", &buf)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	result := collection.FindOne(context.TODO(), bson.D{{"short", "abc"}}, options.FindOne())
	var url URL
	result.Decode(&url)
	assert.Equal(t, url.Long, "https://google.com")
}

func TestShortenRouteFailure(t *testing.T) {
	r := setupRouter()

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(URL{Long: "https://google.com"})
	if err != nil {
		panic(err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/shorten", &buf)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShortenRouteWithPassword(t *testing.T) {
	r := setupRouter()

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(URL{Long: "https://google.com", Short: "def", Password: "password"})
	if err != nil {
		panic(err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/shorten", &buf)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	result := collection.FindOne(context.TODO(), bson.D{{"short", "def"}}, options.FindOne())
	var url URL
	result.Decode(&url)
	assert.Equal(t, url.Long, "https://google.com")
}

func TestRedirectRoute(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/redirect/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Contains(t, w.Body.String(), "google")
}

func TestRedirectRouteWithPassword(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/redirect/def", nil)
	req.Header.Add("Authorization", "bearer password")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Contains(t, w.Body.String(), "google")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/redirect/def", nil)
	req.Header.Add("Authorization", "bearer wrong-password")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRedirectNotFound(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/redirect/hey", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestShortenRouteWithDuplicateShorthand(t *testing.T) {
	r := setupRouter()

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(URL{Long: "https://amazon.com", Short: "abc"})
	if err != nil {
		panic(err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/shorten", &buf)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	result := collection.FindOne(context.TODO(), bson.D{{"short", "abc"}}, options.FindOne())
	var url URL
	result.Decode(&url)
	assert.Equal(t, url.Long, "https://google.com")
}

func TestExpiredRoute(t *testing.T) {
	r := setupRouter()

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(URL{Long: "https://google.com", Short: "hello", ExpiredAt: time.Now().Add(-1 * time.Hour)})
	if err != nil {
		panic(err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/shorten", &buf)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/redirect/hello", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusGone, w.Code)
}

func TestAnalyticsWithoutPeriod(t *testing.T) {
	r := setupRouter()

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(URL{Long: "https://google.com", Short: "test"})
	if err != nil {
		panic(err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/shorten", &buf)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/redirect/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/analytics/test", nil)
	r.ServeHTTP(w, req)

	var resp AnalyticsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	assert.Equal(t, 1, resp.Clicks)
}

func TestAnalyticsWithPeriod(t *testing.T) {
	r := setupRouter()

	var documents []interface{}
	for i := 1; i < 11; i++ {
		documents = append(documents, Click{Short: "TestingPeriod", Time: time.Now().Add(time.Duration(-i*10) * time.Minute)})
	}
	collection.InsertMany(context.TODO(), documents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/analytics/TestingPeriod", nil)

	q := req.URL.Query()
	q.Add("start", time.Now().Add(-25*time.Minute).Format(time.RFC3339))
	q.Add("finish", time.Now().Add(-15*time.Minute).Format(time.RFC3339))
	req.URL.RawQuery = q.Encode()

	r.ServeHTTP(w, req)

	var resp AnalyticsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	assert.Equal(t, 1, resp.Clicks)
}
