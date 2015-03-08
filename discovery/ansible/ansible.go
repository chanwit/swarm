package ansible

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/discovery"
)

type AnsibleDiscoveryService struct {
	heartbeat int
	file      string
	section   string
}

func init() {
	service := &AnsibleDiscoveryService{}
	discovery.Register("ansible", service)
	// alias to aiyara
	discovery.Register("aiyara", service)
}

func inc(s string) string {
	if s == "" {
		return "1"
	}
	i := len(s) - 1
	if s[i] == '9' {
		return inc(s[:len(s)-1]) + "0"

	} else if s[i] == 'Z' {
		return inc(s[:len(s)-1]) + "A"

	} else if s[i] == 'z' {
		return inc(s[:len(s)-1]) + "a"

	} else {
		return s[:len(s)-1] + string(s[i]+1)
	}
	return s
}

func generate(pattern string) []string {
	result := make([]string, 0)

	r, _ := regexp.Compile(`\[(.+):(.+)\]`)
	submatch := r.FindStringSubmatch(pattern)
	template := r.ReplaceAllString(pattern, "%s")

	from := submatch[1]
	to := submatch[2]
	for i := from; ; i = inc(i) {
		each := fmt.Sprintf(template, i)
		result = append(result, each)
		if i == to {
			break
		}
	}
	return result
}

func readSection(data string, section string) ([]string, error) {
	m := map[string][]string{}
	lines := strings.Split(data, "\n")
	currentSection := "all"
	all := make([]string, 0)
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") {
			currentSection = line[1 : len(line)-1]
			log.Info(currentSection)
			m[currentSection] = make([]string, 0)
		} else {
			tokens := strings.Split(line, " ")
			port := "2375"

			if len(tokens) > 1 {
				for _, t := range tokens {
					attr := strings.SplitN(t, "=", 2)
					if attr[0] == "docker_port" {
						port = attr[1]
					}
				}
			}

			// generator
			if strings.Contains(tokens[0], "[") {
				result := generate(tokens[0])
				for _, r := range result {
					if strings.Contains(r, ":") == false {
						r = r + ":" + port
					}
					m[currentSection] = append(m[currentSection], r)
					all = append(all, r)
				}
			} else {
				if strings.Contains(tokens[0], ":") == false {
					tokens[0] = tokens[0] + ":" + port
				}
				m[currentSection] = append(m[currentSection], tokens[0])
				all = append(all, tokens[0])
			}
		}
	}
	if section == "all" {
		return all, nil
	}
	return m[section], nil
}

func (s *AnsibleDiscoveryService) Initialize(path string, heartbeat int) error {
	str := strings.Split(path, "/")
	if len(str) == 2 {
		s.file = "/etc/ansible/hosts"
		s.section = str[1]
	} else if len(str) > 2 {
		s.file = "/" + strings.Join(str[1:len(str)-1], "/")
		s.section = str[len(str)-1]
	}
	s.heartbeat = heartbeat
	return nil
}

func (s *AnsibleDiscoveryService) Fetch() ([]*discovery.Entry, error) {
	data, err := ioutil.ReadFile(s.file)
	if err != nil {
		return nil, err
	}

	hosts, err := readSection(string(data), s.section)
	if err != nil {
		return nil, err
	}
	return discovery.CreateEntries(hosts)
}

func (s *AnsibleDiscoveryService) Watch(callback discovery.WatchCallback) {
	for _ = range time.Tick(time.Duration(s.heartbeat) * time.Second) {
		entries, err := s.Fetch()
		if err == nil {
			callback(entries)
		}
	}
}

func (s *AnsibleDiscoveryService) Register(addr string) error {
	return discovery.ErrNotImplemented
}
