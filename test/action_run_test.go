package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/stretchr/testify/require"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/stretchr/testify/assert"
)

type ActionConfig struct {
	iacName    string
	iacType    string
	iacVersion string
	tgVersion  string
}

func TestTerragruntAction(t *testing.T) {
	t.Parallel()
	tag := buildActionImage(t)
	buildImage(t, "ssh-agent:local", "ssh-agent")

	testCases := []ActionConfig{
		{"Terraform", "TF", "1.4.6", "0.46.3"},
		{"OpenTofu", "TOFU", "1.6.0", "0.53.3"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.iacName, func(t *testing.T) {
			t.Parallel()
			t.Run("testActionIsExecuted", func(t *testing.T) {
				t.Parallel()
				testActionIsExecuted(t, tc, tag)
			})
			t.Run("testActionIsExecutedSSHProject", func(t *testing.T) {
				t.Parallel()
				testActionIsExecutedSSHProject(t, tc, tag)
			})
			t.Run("testOutputPlanIsUsedInApply", func(t *testing.T) {
				t.Parallel()
				testOutputPlanIsUsedInApply(t, tc, tag)
			})
			t.Run("testRunAllIsExecute", func(t *testing.T) {
				t.Parallel()
				testRunAllIsExecuted(t, tc, tag)
			})
			t.Run("testAutoApproveDelete", func(t *testing.T) {
				t.Parallel()
				testAutoApproveDelete(t, tc, tag)
			})
		})
	}
}

func testActionIsExecuted(t *testing.T, actionConfig ActionConfig, tag string) {
	fixturePath := prepareFixture(t, "fixture-action-execution")

	outputTF := runAction(t, actionConfig, false, tag, fixturePath, "plan")
	assert.Contains(t, outputTF, "You can apply this plan to save these new output values to the "+actionConfig.iacName)
}

func testActionIsExecutedSSHProject(t *testing.T, actionConfig ActionConfig, tag string) {
	fixturePath := prepareFixture(t, "fixture-action-execution-ssh")

	outputTF := runAction(t, actionConfig, true, tag, fixturePath, "plan")
	assert.Contains(t, outputTF, "You can apply this plan to save these new output values to the "+actionConfig.iacName)
}

func testOutputPlanIsUsedInApply(t *testing.T, actionConfig ActionConfig, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, actionConfig, false, tag, fixturePath, "run-all plan -out=plan.out")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", actionConfig.iacName)

	output = runAction(t, actionConfig, false, tag, fixturePath, "run-all apply plan.out")
	assert.Contains(t, output, "1 added, 0 changed, 0 destroyed", actionConfig.iacName)
}

func testRunAllIsExecuted(t *testing.T, actionConfig ActionConfig, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, actionConfig, false, tag, fixturePath, "run-all plan")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", actionConfig.iacName)

	output = runAction(t, actionConfig, false, tag, fixturePath, "run-all apply")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", actionConfig.iacName)

	output = runAction(t, actionConfig, false, tag, fixturePath, "run-all destroy")
	assert.Contains(t, output, "0 to add, 0 to change, 1 to destroy", actionConfig.iacName)
	assert.Contains(t, output, "Destroy complete! Resources: 1 destroyed", actionConfig.iacName)
}

func testAutoApproveDelete(t *testing.T, actionConfig ActionConfig, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, actionConfig, false, tag, fixturePath, "run-all plan -out=plan.out")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy")

	output = runAction(t, actionConfig, false, tag, fixturePath, "run-all apply plan.out")
	assert.Contains(t, output, "1 added, 0 changed, 0 destroyed", actionConfig.iacName)

	// run destroy with auto-approve
	output = runAction(t, actionConfig, false, tag, fixturePath, "run-all plan -destroy -out=destroy.out")
	assert.Contains(t, output, "0 to add, 0 to change, 1 to destroy", actionConfig.iacName)

	output = runAction(t, actionConfig, false, tag, fixturePath, "run-all apply -destroy destroy.out")
	assert.Contains(t, output, "Resources: 0 added, 0 changed, 1 destroyed", actionConfig.iacName)
}

func runAction(t *testing.T, actionConfig ActionConfig, sshAgent bool, tag, fixturePath string, command string) string {

	opts := &docker.RunOptions{
		EnvironmentVariables: []string{
			"INPUT_" + actionConfig.iacType + "_VERSION=" + actionConfig.iacVersion,
			"INPUT_TG_VERSION=" + actionConfig.tgVersion,
			"INPUT_TG_COMMAND=" + command,
			"INPUT_TG_DIR=/github/workspace/code",
			"GITHUB_OUTPUT=/tmp/logs",
		},
		Volumes: []string{
			fixturePath + ":/github/workspace/code",
		},
	}

	if sshAgent {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			assert.NoError(t, err)
		}
		sshPath := filepath.Join(homeDir, ".ssh")

		r := random.Random(1, 1000)
		socketPath := fmt.Sprintf("/tmp/ssh-agent.sock.%d", r)
		sshAgentID := docker.RunAndGetID(t, "ssh-agent:local", &docker.RunOptions{
			Detach: true,
			Remove: true,
			EnvironmentVariables: []string{
				"SSH_AUTH_SOCK=" + socketPath,
			},
			Volumes: []string{
				"/tmp:/tmp",
				sshPath + ":/root/keys",
			},
		})
		defer docker.Stop(t, []string{sshAgentID}, &docker.StopOptions{})
		opts.Volumes = append(opts.Volumes, "/tmp:/tmp")
		opts.EnvironmentVariables = append(opts.EnvironmentVariables, "SSH_AUTH_SOCK="+socketPath)
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
