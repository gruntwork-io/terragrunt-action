package test

import (
	"github.com/gruntwork-io/terratest/modules/files"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/stretchr/testify/assert"
)

func TestActionIsExecuted(t *testing.T) {
	tag := buildActionImage(t)

	path, err := files.CopyTerraformFolderToTemp("fixture-action-execution", "test")
	assert.NoError(t, err)

	err = os.Chmod(path, 0777)
	assert.NoError(t, err)

	opts := &docker.RunOptions{
		EnvironmentVariables: []string{
			"INPUT_TF_VERSION=1.4.6",
			"INPUT_TG_VERSION=0.46.3",
			"INPUT_TG_COMMAND=plan",
			"INPUT_TG_DIR=/github/workspace/fixture-action-execution",
			"GITHUB_OUTPUT=/tmp/logs",
		},
		Volumes: []string{path + ":/github/workspace/fixture-action-execution"},
	}
	output := docker.Run(t, tag, opts)
	assert.Contains(t, output, "You can apply this plan to save these new output values to the Terraform")
}
