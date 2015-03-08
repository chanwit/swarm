package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratorNotGenerate(t *testing.T) {
	ips := Generate("127.0.0.1")
	assert.Equal(t, len(ips), 1)
	assert.Equal(t, ips[0], "127.0.0.1")
}

func TestGeneratorWithPortNotGenerate(t *testing.T) {
	ips := Generate("127.0.0.1:8080")
	assert.Equal(t, len(ips), 1)
	assert.Equal(t, ips[0], "127.0.0.1:8080")
}

func TestGeneratorMatchFailedNotGenerate(t *testing.T) {
	ips := Generate("127.0.0.[1]")
	assert.Equal(t, len(ips), 1)
	assert.Equal(t, ips[0], "127.0.0.[1]")
}

func TestGeneratorSimple(t *testing.T) {
	ips := Generate("127.0.0.[00:09]")
	assert.Equal(t, len(ips), 10)
	assert.Equal(t, ips[0], "127.0.0.00")
	assert.Equal(t, ips[1], "127.0.0.01")
	assert.Equal(t, ips[2], "127.0.0.02")
	assert.Equal(t, ips[3], "127.0.0.03")
	assert.Equal(t, ips[4], "127.0.0.04")
	assert.Equal(t, ips[5], "127.0.0.05")
	assert.Equal(t, ips[6], "127.0.0.06")
	assert.Equal(t, ips[7], "127.0.0.07")
	assert.Equal(t, ips[8], "127.0.0.08")
	assert.Equal(t, ips[9], "127.0.0.09")
}

func TestGeneratorWithPort(t *testing.T) {
	ips := Generate("127.0.0.[1:11]:2375")
	assert.Equal(t, len(ips), 11)
	assert.Equal(t, ips[0], "127.0.0.1:2375")
	assert.Equal(t, ips[1], "127.0.0.2:2375")
	assert.Equal(t, ips[2], "127.0.0.3:2375")
	assert.Equal(t, ips[3], "127.0.0.4:2375")
	assert.Equal(t, ips[4], "127.0.0.5:2375")
	assert.Equal(t, ips[5], "127.0.0.6:2375")
	assert.Equal(t, ips[6], "127.0.0.7:2375")
	assert.Equal(t, ips[7], "127.0.0.8:2375")
	assert.Equal(t, ips[8], "127.0.0.9:2375")
	assert.Equal(t, ips[9], "127.0.0.10:2375")
	assert.Equal(t, ips[10], "127.0.0.11:2375")
}

func TestGenerateHostnames(t *testing.T) {
	ips := Generate("web[00:03].abc.com")
	assert.Equal(t, "web00.abc.com", ips[0])
	assert.Equal(t, "web01.abc.com", ips[1])
	assert.Equal(t, "web02.abc.com", ips[2])
	assert.Equal(t, "web03.abc.com", ips[3])

	ips = Generate("[00:03]h.abc.com")
	assert.Equal(t, "00h.abc.com", ips[0])
	assert.Equal(t, "01h.abc.com", ips[1])
	assert.Equal(t, "02h.abc.com", ips[2])
	assert.Equal(t, "03h.abc.com", ips[3])

	ips = Generate("web-[a1z:a2b].example.com")
	assert.Equal(t, "web-a1z.example.com", ips[0])
	assert.Equal(t, "web-a2a.example.com", ips[1])
	assert.Equal(t, "web-a2b.example.com", ips[2])
}
