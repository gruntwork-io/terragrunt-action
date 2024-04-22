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
		iac_name    string
		iac_type    string
		iac_version string
		tg_version  string
	}{
		{"Terraform", "TF", "1.4.6", "0.46.3"},
		{"OpenTofu", "TOFU", "1.6.0", "0.53.3"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.iac_name, func(t *testing.T) {
			t.Parallel()
			t.Run("testActionIsExecuted", func(t *testing.T) {
				t.Parallel()
				testActionIsExecuted(t, tc.iac_type, tc.iac_name, tc.iac_version, tc.tg_version, tag)
			})
			t.Run("testOutputPlanIsUsedInApply", func(t *testing.T) {
				t.Parallel()
				testOutputPlanIsUsedInApply(t, tc.iac_type, tc.iac_name, tc.iac_version, tc.tg_version, tag)
			})
			t.Run("testRunAllIsExecute", func(t *testing.T) {
				t.Parallel()
				testRunAllIsExecuted(t, tc.iac_type, tc.iac_name, tc.iac_version, tc.tg_version, tag)
			})
			t.Run("testAutoApproveDelete", func(t *testing.T) {
				t.Parallel()
				testAutoApproveDelete(t, tc.iac_type, tc.iac_name, tc.iac_version, tc.tg_version, tag)
			})
		})
	}
}

func testActionIsExecuted(t *testing.T, iac_type string, iac_name string, iac_version string, tg_version string, tag string) {
	fixturePath := prepareFixture(t, "fixture-action-execution")

	outputTF := runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "plan")
	assert.Contains(t, outputTF, "You can apply this plan to save these new output values to the "+iac_name)
}

func testOutputPlanIsUsedInApply(t *testing.T, iac_type string, iac_name string, iac_version string, tg_version string, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "run-all plan -out=plan.out")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy")

	output = runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "run-all apply plan.out")
	assert.Contains(t, output, "1 added, 0 changed, 0 destroyed")
}

func testRunAllIsExecuted(t *testing.T, iac_type string, iac_name string, iac_version string, tg_version string, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "run-all plan")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy")

	output = runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "run-all apply")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy")

	output = runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "run-all destroy")
	assert.Contains(t, output, "0 to add, 0 to change, 1 to destroy")
	assert.Contains(t, output, "Destroy complete! Resources: 1 destroyed")
}

func testAutoApproveDelete(t *testing.T, iac_type string, iac_name string, iac_version string, tg_version string, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "run-all plan -out=plan.out")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy")

	output = runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "run-all apply plan.out")
	assert.Contains(t, output, "1 added, 0 changed, 0 destroyed")

	// run destroy with auto-approve
	output = runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "run-all plan -destroy -out=destroy.out")
	assert.Contains(t, output, "0 to add, 0 to change, 1 to destroy")

	output = runAction(t, tag, fixturePath, iac_type, iac_version, tg_version, "run-all apply -destroy destroy.out")
	assert.Contains(t, output, "Resources: 0 added, 0 changed, 1 destroyed")
}

func runAction(t *testing.T, tag, fixturePath, iac_type string, iac_version string, tg_version string, command string) string {
	opts := &docker.RunOptions{
		EnvironmentVariables: []string{
			"INPUT_" + iac_type + "_VERSION=" + iac_version,
			"INPUT_TG_VERSION=" + tg_version,
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
