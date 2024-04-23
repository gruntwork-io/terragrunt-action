package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/random"
)

func buildActionImage(t *testing.T) string {
	tag := "terragrunt-action:" + random.UniqueId()
	buildOptions := &docker.BuildOptions{
		Tags: []string{tag},
	}
	docker.Build(t, "..", buildOptions)
	return tag
}
