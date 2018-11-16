package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/dustin/go-humanize"
)

func _http(method string, url string, headers map[string]string, body []byte, expectedStatus int) (*http.Response, error) {
	start := time.Now()
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(responseBody))

	log.Printf("%s %s ... %s (%s; %s)", req.Method, req.URL.String(), resp.Status, time.Since(start), humanize.Bytes(uint64(len(responseBody))))

	if expectedStatus != -1 && resp.StatusCode != expectedStatus {
		return resp, fmt.Errorf("unexpected status: %d (expected %d)", resp.StatusCode, expectedStatus)
	}

	return resp, nil
}

func _get(url string, headers map[string]string, expectedStatus int) (*http.Response, error) {
	return _http(
		"GET",
		url,
		headers,
		[]byte{},
		expectedStatus,
	)
}

// GoogleMapsPolylineURL returns url for generating a static image file using Google Static Maps API
func GoogleMapsPolylineURL(polyline string, mapID string, key string) (*url.URL, error) {
	u, err := url.Parse("https://maps.googleapis.com/maps/api/staticmap")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("size", "640x640")
	q.Set("scale", "2")
	//q.Set("maptype", "terrain")
	q.Set("map_id", mapID)
	q.Set("path", "weight:3|color:red|enc:"+polyline)
	q.Set("key", key)
	u.RawQuery = q.Encode()
	return u, nil
}

// APIGoogleStaticMaps returns []byte representing a PNG file of the image or error
func APIGoogleStaticMaps(apiKey, polyline string, mapID string) ([]byte, error) {
	url, err := GoogleMapsPolylineURL(polyline, mapID, apiKey)
	if err != nil {
		return nil, err
	}
	resp, err := _get(
		url.String(),
		map[string]string{},
		-1,
	)
	if err != nil {
		return nil, err
	}
	pngBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return pngBytes, nil
}
