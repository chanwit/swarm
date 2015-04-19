package cluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRulesWithDefinedKernels(t *testing.T) {
	yamlRules := `
kernels:
  cubie\w+$: arm
  default: amd64

^(nginx)$:
  amd64: $1
  default: aiyara/$1:latest.$arch
`
	kernelMap, _ := parseRules(yamlRules)
	assert.Equal(t, kernelMap[`cubie\w+$`], `arm`)
	assert.Equal(t, kernelMap[`default`], `amd64`)
}

func TestMatchKernelToArch(t *testing.T) {
	yamlRules := `
kernels:
  cubie\w+$: arm
  default: amd64

^(nginx)$:
  amd64: $1
  default: aiyara/$1:latest.$arch
`
	assert.Equal(t, arch0(yamlRules, "3.4.106-cubieboard"), `arm`)
	assert.Equal(t, arch0(yamlRules, "3.4.106-cubietruck"), `arm`)
	assert.Equal(t, arch0(yamlRules, "3.4.106-cubie"), `amd64`)
}

func TestMatchKernelToArchUnsorted(t *testing.T) {
	yamlRules := `
kernels:
  default: amd64
  cubie\w+$: arm

^(nginx)$:
  amd64: $1
  default: aiyara/$1:latest.$arch
`
	assert.Equal(t, arch0(yamlRules, "3.4.106-cubieboard"), `arm`)
	assert.Equal(t, arch0(yamlRules, "3.4.106-cubie"), `amd64`)
}

func TestParseRules(t *testing.T) {
	yamlRules := `
^(nginx)$:
  amd64: $1
  default: aiyara/$1:latest.$arch
`
	_, ruleMap := parseRules(yamlRules)
	assert.Equal(t, ruleMap[`^(nginx)$`]["amd64"], `$1`)
	assert.Equal(t, ruleMap[`^(nginx)$`]["default"], `aiyara/$1:latest.$arch`)
}

func TestParseRules_2(t *testing.T) {
	yamlRules := `
"^(nginx):(latest)$":
  amd64: $1
  default: aiyara/$1:latest.$arch
`
	_, ruleMap := parseRules(yamlRules)
	assert.Equal(t, ruleMap[`^(nginx):(latest)$`]["amd64"], `$1`)
	assert.Equal(t, ruleMap[`^(nginx):(latest)$`]["default"], `aiyara/$1:latest.$arch`)
}

func TestPullImageSimpleRewrite_01(t *testing.T) {
	yamlRules := `
^(nginx)$:
  amd64: $1_2
  default: aiyara/$1:latest.$arch
`
	_, ruleMap := parseRules(yamlRules)
	var labels map[string]string
	var result string

	labels = map[string]string{"architecture": "amd64"}
	result, _ = rewrite0("nginx", labels, ruleMap)
	assert.Equal(t, result, "nginx_2")
}

func TestPullImageSimpleRewriteReuseGroup(t *testing.T) {
	yamlRules := `
^(nginx)$:
  amd64: $1_$1_2
  default: aiyara/$1:latest.$arch
`
	_, ruleMap := parseRules(yamlRules)
	var labels map[string]string
	var result string

	labels = map[string]string{"architecture": "amd64"}
	result, _ = rewrite0("nginx", labels, ruleMap)
	assert.Equal(t, result, "nginx_nginx_2")
}

func TestPullImageSimpleRewrite_Rules_not_match(t *testing.T) {
	yamlRules := `
"^(nginx):(latest)$":
  amd64: $1_$2_3
  default: aiyara/$1:latest.$arch
`
	_, ruleMap := parseRules(yamlRules)
	var labels map[string]string
	var result string
	var err error

	// No label existed
	labels = map[string]string{}
	result, err = rewrite0("nginx", labels, ruleMap)
	assert.Error(t, err) // err not match
	assert.Equal(t, result, "nginx")

	// Not match
	labels = map[string]string{"architecture": "amd64"}
	result, err = rewrite0("nginx", labels, ruleMap)
	assert.Error(t, err)
	assert.Equal(t, result, "nginx")
}

func TestPullImageRewriteMultipleGroups(t *testing.T) {
	yamlRules := `
"^(nginx):(latest)$":
  amd64: $1_$2_3
  default: aiyara/$1:latest.$arch
`
	_, ruleMap := parseRules(yamlRules)
	var labels map[string]string
	var result string
	var err error

	labels = map[string]string{"architecture": "amd64"}
	result, err = rewrite0("nginx:latest", labels, ruleMap)
	assert.NoError(t, err)
	assert.Equal(t, result, "nginx_latest_3")
}

func TestPullImageRewriteDefaultCase(t *testing.T) {
	yamlRules := `
^(nginx)$:
  amd64: $1_2
  default: aiyara/$1:latest.$arch
`
	_, ruleMap := parseRules(yamlRules)
	var labels map[string]string
	var result string
	var err error

	labels = map[string]string{"architecture": "amd64"}
	result, err = rewrite0("nginx", labels, ruleMap)
	assert.NoError(t, err)
	assert.Equal(t, result, "nginx_2")

	labels = map[string]string{"architecture": "arm"}
	result, err = rewrite0("nginx", labels, ruleMap)
	assert.NoError(t, err)
	assert.Equal(t, result, "aiyara/nginx:latest.arm")

	labels = map[string]string{"architecture": "386"}
	result, err = rewrite0("nginx", labels, ruleMap)
	assert.NoError(t, err)
	assert.Equal(t, result, "aiyara/nginx:latest.386")

	labels = map[string]string{"architecture": "arm64"}
	result, err = rewrite0("nginx", labels, ruleMap)
	assert.NoError(t, err)
	assert.Equal(t, result, "aiyara/nginx:latest.arm64")

}

func TestHardCodedRules(t *testing.T) {
	ruleMap := parseRules0(DefaultRules)
	var labels map[string]string
	var result string
	var err error

	fmt.Printf(">>> %s\n", ruleMap)

	// no label['architecture'] so default to amd64
	labels = map[string]string{}
	result, err = rewrite0("nginx", labels, ruleMap[2])
	assert.NoError(t, err)
	assert.Equal(t, result, "nginx")

	result, err = rewrite0("nginx:latest", labels, ruleMap[1])
	assert.NoError(t, err)
	assert.Equal(t, result, "nginx:latest")

	result, err = rewrite0("my/nginx:latest", labels, ruleMap[0])
	assert.NoError(t, err)
	assert.Equal(t, result, "my/nginx:latest")
}

func TestRewriteWithDefaultRules(t *testing.T) {
	engine := &Engine{}

	// no label['architecture'] existed, use amd64
	image := "nginx"
	engine.Labels = map[string]string{}
	assert.Equal(t, "nginx", rewriteWithDefaultRules(engine, image))

	// label['architecture'] = amd64, so should rewrite with amd64
	image = "nginx"
	engine.Labels = map[string]string{"architecture": "amd64"}
	assert.Equal(t, "nginx", rewriteWithDefaultRules(engine, image))

	// label['architecture'] = arm, goes with the arm rule
	image = "nginx"
	engine.Labels = map[string]string{"architecture": "arm"}
	assert.Equal(t, "aiyara/nginx:latest.arm", rewriteWithDefaultRules(engine, image))

	// label['architecture'] = 386, goes with the 386 rule
	image = "nginx"
	engine.Labels = map[string]string{"architecture": "386"}
	assert.Equal(t, "aiyara/nginx:latest.386", rewriteWithDefaultRules(engine, image))

	// label['architecture'] = arm, and the image contains tag
	image = "nginx:tag"
	engine.Labels = map[string]string{"architecture": "arm"}
	assert.Equal(t, "aiyara/nginx:tag.arm", rewriteWithDefaultRules(engine, image))

	// label['architecture'] = arm, and the image contains repository name and tag
	// rewrite with underscore
	image = "chanwit/zookeeper:3.4.6"
	engine.Labels = map[string]string{"architecture": "arm"}
	assert.Equal(t, "aiyara/chanwit_zookeeper:3.4.6.arm", rewriteWithDefaultRules(engine, image))
}
