package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatRecording(t *testing.T) {
	stats := Stats{}
	request := Request{
		Server:   "Simple server",
		Endpoint: "/some_url",
		Method:   "POST",
	}
	stats.Add(request)
	assert.Equal(t, stats.Requests, []Request{Request{
		Server:   "Simple server",
		Endpoint: "/some_url",
		Method:   "POST",
	}})
}

func TestStatFetcher(t *testing.T) {
	stats := Stats{}
	request := Request{
		Server:   "Simple server",
		Endpoint: "/some_url",
		Method:   "POST",
	}
	stats.Add(request)
	stats.Add(request)
	assert.Equal(t, stats.Get(request), 2)
}
