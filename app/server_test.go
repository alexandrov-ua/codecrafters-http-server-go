package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseHeader1(t *testing.T) {
	name, val := ParseHeaderLine([]byte("Accept-Encoding: invalid-encoding-1, gzip, invalid-encoding-2"))

	assert.Equal(t, "Accept-Encoding", name)
	assert.Len(t, val, 3)
	assert.Contains(t, val, "invalid-encoding-1")
	assert.Contains(t, val, "invalid-encoding-2")
	assert.Contains(t, val, "gzip")
}

func Test_ParseHeader2(t *testing.T) {
	name, val := ParseHeaderLine([]byte("Accept: text/html"))
	assert.Equal(t, "Accept", name)

	assert.Contains(t, val, "text/html")

}

func Test_ParseHeader3(t *testing.T) {
	name, val := ParseHeaderLine([]byte("x-xss-protection: 0"))
	assert.Equal(t, "x-xss-protection", name)

	assert.Len(t, val, 1)
	assert.Contains(t, val, "0")

}

func Test_ParseHeader4(t *testing.T) {
	name, val := ParseHeaderLine([]byte("x-xss-protection:0"))
	assert.Equal(t, "x-xss-protection", name)

	assert.Len(t, val, 1)
	assert.Contains(t, val, "0")

}
