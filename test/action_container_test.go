package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/stretchr/testify/assert"
)

func TestActionContainerIsBuilt(t *testing.T) {
	tag := buildActionImage(t)
	
	opts := &docker.RunOptions{Entrypoint: "/bin/bash", Command: []string{"-c", "ls /action"}}
	output := docker.Run(t, tag, opts)
	assert.Equal(t, "main.sh", output)
}
