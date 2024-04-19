package test

import (
	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/random"
	"testing"
)

func buildActionImage(t *testing.T) string {
	tag := "terragrunt-action:" + random.UniqueId()
	buildOptions := &docker.BuildOptions{
		Architectures: []string{"linux/amd64"},
		Tags: []string{tag},
	}
	docker.Build(t, "..", buildOptions)
	return tag
}
