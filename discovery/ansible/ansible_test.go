package ansible

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	discovery := &AnsibleDiscoveryService{}
	discovery.Initialize("/path/to/file/section", 0)
	assert.Equal(t, discovery.file, "/path/to/file")
	assert.Equal(t, discovery.section, "section")
}

func TestInitialize2(t *testing.T) {
	discovery := &AnsibleDiscoveryService{}
	discovery.Initialize("/section", 0)
	assert.Equal(t, discovery.file, "/etc/ansible/hosts")
	assert.Equal(t, discovery.section, "section")
}

func TestAllReadSection(t *testing.T) {
	hosts, err := readSection(`
[web]
192.168.0.1

[db]
192.168.0.2`, "all")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(hosts))
	assert.Equal(t, "192.168.0.1:2375", hosts[0])
	assert.Equal(t, "192.168.0.2:2375", hosts[1])
}

func TestReadSection(t *testing.T) {
	data := `
# This is comment

[web]
192.168.0.1
# This is comment
192.168.0.3

[db]
# This is comment
192.168.0.2

[with_port]
# This is comment
192.168.0.4 swarm_port=2375
`
	web, err := readSection(data, "web")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(web))
	assert.Equal(t, "192.168.0.1:2375", web[0])
	assert.Equal(t, "192.168.0.3:2375", web[1])

	db, err := readSection(data, "db")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(db))
	assert.Equal(t, "192.168.0.2:2375", db[0])

	with_port, err := readSection(data, "with_port")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(with_port))
	assert.Equal(t, "192.168.0.4:2375", with_port[0])
}

func TestReadGenerator(t *testing.T) {
	data := `
[web]
# Support generator
192.168.0.[1:5]
`
	web, err := readSection(data, "web")
	assert.NoError(t, err)
	assert.Equal(t, 5, len(web))
	assert.Equal(t, "192.168.0.1:2375", web[0])
	assert.Equal(t, "192.168.0.2:2375", web[1])
	assert.Equal(t, "192.168.0.3:2375", web[2])
	assert.Equal(t, "192.168.0.4:2375", web[3])
	assert.Equal(t, "192.168.0.5:2375", web[4])
}

func TestGenerate(t *testing.T) {
	ip := generate("1.2.3.[4:6]")
	assert.Equal(t, "1.2.3.4", ip[0])
	assert.Equal(t, "1.2.3.5", ip[1])
	assert.Equal(t, "1.2.3.6", ip[2])

	ip = generate("1.2.3.[09:11]")
	assert.Equal(t, "1.2.3.09", ip[0])
	assert.Equal(t, "1.2.3.10", ip[1])
	assert.Equal(t, "1.2.3.11", ip[2])

	ip = generate("1.2.3.[9:11]")
	assert.Equal(t, "1.2.3.9", ip[0])
	assert.Equal(t, "1.2.3.10", ip[1])
	assert.Equal(t, "1.2.3.11", ip[2])

	ip = generate("web[00:03].abc.com")
	assert.Equal(t, "web00.abc.com", ip[0])
	assert.Equal(t, "web01.abc.com", ip[1])
	assert.Equal(t, "web02.abc.com", ip[2])
	assert.Equal(t, "web03.abc.com", ip[3])

	ip = generate("[00:03]h.abc.com")
	assert.Equal(t, "00h.abc.com", ip[0])
	assert.Equal(t, "01h.abc.com", ip[1])
	assert.Equal(t, "02h.abc.com", ip[2])
	assert.Equal(t, "03h.abc.com", ip[3])

	ip = generate("web-[a1z:a2b].example.com")
	assert.Equal(t, "web-a1z.example.com", ip[0])
	assert.Equal(t, "web-a2a.example.com", ip[1])
	assert.Equal(t, "web-a2b.example.com", ip[2])
}

func TestRegister(t *testing.T) {
	discovery := &AnsibleDiscoveryService{file: "/path/to/file", section: "all"}
	assert.Error(t, discovery.Register("0.0.0.0"))
}
