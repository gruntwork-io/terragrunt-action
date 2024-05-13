package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		{"Terraform1.5", "TF", "1.5.7", "0.55.18"},
		{"Terraform1.7", "TF", "1.7.5", "0.55.18"},
		{"Terraform1.8", "TF", "1.8.3", "0.55.18"},
		{"OpenTofu1.6", "TOFU", "1.6.0", "0.55.18"},
		{"OpenTofu1.7", "TOFU", "1.7.0", "0.55.18"},
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
			t.Run("testGitWorkingAction", func(t *testing.T) {
				t.Parallel()
				testGitWorkingAction(t, tc, tag)
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
	assert.Contains(t, outputTF, "You can apply this plan to save these new output values to the "+fetchIacType(actionConfig))
}

func testActionIsExecutedSSHProject(t *testing.T, actionConfig ActionConfig, tag string) {
	fixturePath := prepareFixture(t, "fixture-action-execution-ssh")

	outputTF := runAction(t, actionConfig, true, tag, fixturePath, "plan")
	assert.Contains(t, outputTF, "You can apply this plan to save these new output values to the "+fetchIacType(actionConfig))
}

func testOutputPlanIsUsedInApply(t *testing.T, actionConfig ActionConfig, tag string) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, actionConfig, false, tag, fixturePath, "run-all plan -out=plan.out --terragrunt-log-level debug")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", actionConfig.iacName)

	output = runAction(t, actionConfig, false, tag, fixturePath, "run-all apply plan.out --terragrunt-log-level debug")
	assert.Contains(t, output, "1 added, 0 changed, 0 destroyed", actionConfig.iacName)
}

func testGitWorkingAction(t *testing.T, actionConfig ActionConfig, tag string) {
	fixturePath := prepareFixture(t, "fixture-git-commands")
	// init git repo in fixture path, run git init
	_, err := exec.Command("git", "init", fixturePath).CombinedOutput()
	if err != nil {
		t.Fatalf("Error initializing git repo: %v", err)
	}

	output := runAction(t, actionConfig, true, tag, fixturePath, "run-all plan -out=plan.out --terragrunt-log-level debug")
	assert.Contains(t, output, fetchIacType(actionConfig)+" has been successfully initialized!", actionConfig.iacName)
	assert.Contains(t, output, "execute_INPUT_POST_EXEC_1", actionConfig.iacName)
	assert.Contains(t, output, "execute_INPUT_PRE_EXEC_1", actionConfig.iacName)
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

	// check that fixturePath can removed recursively
	err := os.RemoveAll(fixturePath)
	assert.NoError(t, err)
}

func runAction(t *testing.T, actionConfig ActionConfig, sshAgent bool, tag, fixturePath string, command string) string {

	logId := random.Random(1, 5000)
	opts := &docker.RunOptions{
		EnvironmentVariables: []string{
			"INPUT_" + actionConfig.iacType + "_VERSION=" + actionConfig.iacVersion,
			"INPUT_TG_VERSION=" + actionConfig.tgVersion,
			"INPUT_TG_COMMAND=" + command,
			"INPUT_TG_DIR=/github/workspace",
			"INPUT_PRE_EXEC_1=echo 'execute_INPUT_PRE_EXEC_1'",
			"INPUT_POST_EXEC_1=echo 'execute_INPUT_POST_EXEC_1'",
			fmt.Sprintf("GITHUB_OUTPUT=/tmp/github-action-logs.%d", logId),
		},
		Volumes: []string{
			fixturePath + ":/github/workspace",
		},
	}

	// start ssh-agent container with SSH keys to allow clones over SSH
	if sshAgent {
		homeDir, err := os.UserHomeDir()
		assert.NoError(t, err)
		sshPath := filepath.Join(homeDir, ".ssh")

		socketId := random.Random(1, 5000)
		socketPath := fmt.Sprintf("/tmp/ssh-agent.sock.%d", socketId)
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
	return path
}

func fetchIacType(actionConfig ActionConfig) string {
	// return Terraform if OpenTofu based on iacName value
	if strings.ToLower(actionConfig.iacType) == "tf" {
		return "Terraform"
	}
	return "OpenTofu"

}
