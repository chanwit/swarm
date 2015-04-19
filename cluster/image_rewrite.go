package cluster

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	DefaultRules = `
- ^([\w\.]+)/([\w\.]+):([\w\.]+)$:
    amd64: $1/$2:$3
    default: aiyara/$1_$2:$3.$arch

- ^([\w\.]+):([\w\.]+)$:
    amd64: $1:$2
    default: aiyara/$1:$2.$arch

- ^([\w\.]+)$:
    amd64: $1
    default: aiyara/$1:latest.$arch
`
)

var DefaultRuleMap = parseRules0(DefaultRules)

func homepath(p string) string {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	return filepath.Join(home, p)
}

func parseRules0(yamlRules string) []map[string]map[string]string {
	rules := make([]map[string]map[string]string, 0)
	err := yaml.Unmarshal([]byte(yamlRules), &rules)
	if err != nil {
		return nil
	}

	log.Debugf("%s", rules)
	return rules
}

func parseRules(yamlRules string) (map[string]string, map[string]map[string]string) {
	rules := make(map[string]map[string]string)
	err := yaml.Unmarshal([]byte(yamlRules), &rules)
	if err != nil {
		return nil, nil
	}
	kernelMap := rules["kernels"]
	delete(rules, "kernels")
	return kernelMap, rules
}

// Guess architecture from KernelVersion
// Return value defined in rewrite_rules.yaml
// The default value is "amd64"
func arch(kernelVersion string) string {
	filename := filepath.Join(homepath(".swarm"), "rewrite_rules.yaml")
	yamlRules, err := ioutil.ReadFile(filename)
	if err != nil {
		return "amd64"
	}

	return arch0(string(yamlRules), kernelVersion)
}

func arch0(yamlRules string, kernelVersion string) string {
	kernelMap, _ := parseRules(yamlRules)
	for k, v := range kernelMap {
		if k == "default" {
			continue
		}

		re, err := regexp.Compile(k)
		if err != nil {
			continue
		}

		if re.MatchString(kernelVersion) {
			return v
		}
	}

	if def, ok := kernelMap["default"]; ok {
		return def
	}

	return "amd64"
}

func rewrite(e *Engine, image string) string {
	filename := filepath.Join(homepath(".swarm"), "rewrite_rules.yaml")
	yamlRules, err := ioutil.ReadFile(filename)
	if err == nil {
		_, rules := parseRules(string(yamlRules))
		image, _ = rewrite0(image, e.Labels, rules)
		log.Debugf(">>> image name after rewrite: %s", image)
	}
	return image
}

func rewriteWithDefaultRules(e *Engine, image string) string {
	// Already rewrote, skip
	// It's a fix for Compose asking image name more than once
	if string.HasSuffix(image, ".arm") {
		return image
	}
	if string.HasSuffix(image, ".386") {
		return image
	}
	if string.HasSuffix(image, ".amd64") {
		return image
	}

	for _, ruleMap := range DefaultRuleMap {
		result, err := rewrite0(image, e.Labels, ruleMap)
		if err == nil {
			return result
		}
	}
	return image
}

func rewrite0(s string, labels map[string]string, rules map[string]map[string]string) (string, error) {
	// If the label 'architecture' not found
	// we would assume it as 'amd64'
	arch, ok := labels["architecture"]
	if !ok {
		arch = "amd64"
	}

	for k, v := range rules {
		log.Debugf("k = %s", k)
		re, err := regexp.Compile(k)
		if err != nil {
			return s, err
		}

		target := v[arch]
		if target == "" {
			target = v["default"]

			// no default rule defined
			if target == "" {
				continue
			}
		}

		// replace special vars, e.g., $arch ...
		r := strings.NewReplacer("$arch", arch)
		target = r.Replace(target)

		match := re.FindAllStringSubmatch(s, -1)

		// if not match, skip to the next rule
		if len(match) == 0 {
			continue
		}

		for i := 1; i < len(match[0]); i++ {
			target = strings.Replace(target, fmt.Sprintf("$%d", i), match[0][i], -1)
			// log.Infof("match[0][i] = %s", match[0][i])
			// log.Infof("target = %s", target)
		}
		log.Debugf("Rewrote %s as %s", s, target)
		return target, nil
	}

	return s, fmt.Errorf("Rules not match")
}
