package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/stretchr/testify/require"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/stretchr/testify/assert"
)

func TestTerragruntAction(t *testing.T) {
	t.Parallel()
	tag := buildActionImage(t)

	testCases := []struct {
		iacName    string
		iacType    string
		iacVersion string
		tgVersion  string
	}{
		{"Terraform", "TF", "1.4.6", "0.46.3"},
		{"OpenTofu", "TOFU", "1.6.0", "0.53.3"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.iacName, func(t *testing.T) {
			t.Parallel()
			t.Run("testActionIsExecuted", func(t *testing.T) {
				t.Parallel()
				testActionIsExecuted(t, tc.iacType, tc.iacName, tc.iacVersion, tc.tgVersion, tag)
			})
			t.Run("testActionIsExecutedSSHProject", func(t *testing.T) {
				t.Parallel()
				testActionIsExecutedSSHProject(t, tc.iacType, tc.iacName, tc.iacVersion, tc.tgVersion, tag)
			})
			t.Run("testOutputPlanIsUsedInApply", func(t *testing.T) {
				t.Parallel()
				testOutputPlanIsUsedInApply(t, tc.iacType, tc.iacName, tc.iacVersion, tc.tgVersion, tag)
			})
			t.Run("testRunAllIsExecute", func(t *testing.T) {
				t.Parallel()
				testRunAllIsExecuted(t, tc.iacType, tc.iacName, tc.iacVersion, tc.tgVersion, tag)
			})
			t.Run("testAutoApproveDelete", func(t *testing.T) {
				t.Parallel()
				testAutoApproveDelete(t, tc.iacType, tc.iacName, tc.iacVersion, tc.tgVersion, tag)
			})
		})
	}
}

func testActionIsExecuted(t *testing.T, iacType string, iacName string, iacVersion string, tgVersion string, tag string) {
	fixturePath := prepareFixture(t, "fixture-action-execution")

	outputTF := runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "plan")
	assert.Contains(t, outputTF, "You can apply this plan to save these new output values to the "+iacName)
}

func testActionIsExecutedSSHProject(t *testing.T, iacType string, iacName string, iacVersion string, tgVersion string, tag string) {
	fixturePath := prepareFixture(t, "fixture-action-execution-ssh")

	outputTF := runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "plan")
	assert.Contains(t, outputTF, "You can apply this plan to save these new output values to the "+iacName)
}

func testOutputPlanIsUsedInApply(t *testing.T, iacType string, iacName string, iacVersion string, tgVersion string, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "run-all plan -out=plan.out")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", iacName)

	output = runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "run-all apply plan.out")
	assert.Contains(t, output, "1 added, 0 changed, 0 destroyed", iacName)
}

func testRunAllIsExecuted(t *testing.T, iacType string, iacName string, iacVersion string, tgVersion string, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "run-all plan")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", iacName)

	output = runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "run-all apply")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", iacName)

	output = runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "run-all destroy")
	assert.Contains(t, output, "0 to add, 0 to change, 1 to destroy", iacName)
	assert.Contains(t, output, "Destroy complete! Resources: 1 destroyed", iacName)
}

func testAutoApproveDelete(t *testing.T, iacType string, iacName string, iacVersion string, tgVersion string, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "run-all plan -out=plan.out")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy")

	output = runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "run-all apply plan.out")
	assert.Contains(t, output, "1 added, 0 changed, 0 destroyed", iacName)

	// run destroy with auto-approve
	output = runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "run-all plan -destroy -out=destroy.out")
	assert.Contains(t, output, "0 to add, 0 to change, 1 to destroy", iacName)

	output = runAction(t, tag, fixturePath, iacType, iacVersion, tgVersion, "run-all apply -destroy destroy.out")
	assert.Contains(t, output, "Resources: 0 added, 0 changed, 1 destroyed", iacName)
}

func runAction(t *testing.T, tag, fixturePath, iacType string, iacVersion string, tgVersion string, command string) string {
	opts := &docker.RunOptions{
		EnvironmentVariables: []string{
			"INPUT_" + iacType + "_VERSION=" + iacVersion,
			"INPUT_tgVersion=" + tgVersion,
			"INPUT_TG_COMMAND=" + command,
			"INPUT_TG_DIR=/github/workspace/code",
			"GITHUB_OUTPUT=/tmp/logs",
		},
		Volumes: []string{
			fixturePath + ":/github/workspace/code",
		},
		Remove: true,
	}
	return docker.Run(t, tag, opts)
}

func prepareFixture(t *testing.T, fixtureDir string) string {
	path, err := files.CopyTerraformFolderToTemp(fixtureDir, "test")
	require.NoError(t, err)
	// chmod recursive for docker run

	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chmod(path, 0777)
	})
	require.NoError(t, err)

	err = os.Chmod(path, 0777)
	require.NoError(t, err)
	return path
}
