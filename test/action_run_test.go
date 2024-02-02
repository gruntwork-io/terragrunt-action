package test

import (
	"github.com/gruntwork-io/terratest/modules/files"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/stretchr/testify/assert"
)

func TestActionIsExecuted(t *testing.T) {
	tag := buildActionImage(t)
	fixturePath := prepareFixture(t, "fixture-action-execution")

	output := runAction(t, tag, fixturePath, "plan")
	assert.Contains(t, output, "You can apply this plan to save these new output values to the Terraform")
}

func TestOutputPlanIsUsedInApply(t *testing.T) {
	tag := buildActionImage(t)
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, tag, fixturePath, "run-all plan -out=plan.out")
	assert.Contains(t, output, "Plan: 1 to add, 0 to change, 0 to destroy")

	output = runAction(t, tag, fixturePath, "run-all apply plan.out")
	assert.Contains(t, output, "Apply complete! Resources: 1 added, 0 changed, 0 destroyed")
}

func runAction(t *testing.T, tag, fixturePath, command string) string {
	opts := &docker.RunOptions{
		EnvironmentVariables: []string{
			"INPUT_TF_VERSION=1.4.6",
			"INPUT_TG_VERSION=0.46.3",
			"INPUT_TG_COMMAND=" + command,
			"INPUT_TG_DIR=/github/workspace/code",
			"GITHUB_OUTPUT=/tmp/logs",
		},
		Volumes: []string{fixturePath + ":/github/workspace/code"},
	}
	return docker.Run(t, tag, opts)
}

func prepareFixture(t *testing.T, fixtureDir string) string {
	path, err := files.CopyTerraformFolderToTemp(fixtureDir, "test")
	assert.NoError(t, err)
	// chmod recursive for docker run

	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chmod(path, 0777)
	})
	assert.NoError(t, err)

	err = os.Chmod(path, 0777)
	assert.NoError(t, err)
	return path
}
