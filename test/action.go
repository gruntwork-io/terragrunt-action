package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/random"
)

func buildImage(t *testing.T, tag, path string) {
	buildOptions := &docker.BuildOptions{
		Tags: []string{tag},
	}
	docker.Build(t, path, buildOptions)
}

func buildActionImage(t *testing.T) string {
	tag := "terragrunt-action:" + random.UniqueId()
	buildImage(t, tag, "..")
	return tag
}
