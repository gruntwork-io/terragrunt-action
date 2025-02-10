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
	tgDir      string
	tag        string
}

func action(name, iacType, iacVersion, tgVersion, tag string) ActionConfig {

	return ActionConfig{
		iacName:    name,
		iacType:    iacType,
		iacVersion: iacVersion,
		tgVersion:  tgVersion,
		tag:        tag,
	}
}

func TestTerragruntAction(t *testing.T) {
	t.Parallel()
	tag := buildActionImage(t)
	buildImage(t, "ssh-agent:local", "ssh-agent")

	testCases := []ActionConfig{
		action("Terraform1.5", "TF", "1.5.7", "0.55.18", tag),
		action("Terraform1.7", "TF", "1.7.5", "0.55.18", tag),
		action("Terraform1.8", "TF", "1.8.3", "0.55.18", tag),
		action("OpenTofu1.6", "TOFU", "1.6.0", "0.55.18", tag),
		action("OpenTofu1.7", "TOFU", "1.7.0", "0.55.18", tag),
		action("OpenTofu1.8", "TOFU", "1.8.0", "0.72.9", tag),
		action("OpenTofu1.9", "TOFU", "1.9.0", "0.72.9", tag),
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.iacName, func(t *testing.T) {
			t.Parallel()
			t.Run("testActionIsExecuted", func(t *testing.T) {
				t.Parallel()
				testActionIsExecuted(t, tc)
			})
			t.Run("testActionIsExecutedSSHProject", func(t *testing.T) {
				t.Parallel()
				testActionIsExecutedSSHProject(t, tc)
			})
			t.Run("testOutputPlanIsUsedInApply", func(t *testing.T) {
				t.Parallel()
				testOutputPlanIsUsedInApply(t, tc)
			})
			t.Run("testGitWorkingAction", func(t *testing.T) {
				t.Parallel()
				testGitWorkingAction(t, tc)
			})
			t.Run("testRunAllIsExecute", func(t *testing.T) {
				t.Parallel()
				testRunAllIsExecuted(t, tc)
			})
			t.Run("testAutoApproveDelete", func(t *testing.T) {
				t.Parallel()
				testAutoApproveDelete(t, tc)
			})
		})
	}
}

func testActionIsExecuted(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-action-execution")

	outputTF := runAction(t, actionConfig, false, fixturePath, "plan")
	assert.Contains(t, outputTF, "You can apply this plan to save these new output values to the "+fetchIacType(actionConfig))
}

func testActionIsExecutedSSHProject(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-action-execution-ssh")

	outputTF := runAction(t, actionConfig, true, fixturePath, "plan")
	assert.Contains(t, outputTF, "You can apply this plan to save these new output values to the "+fetchIacType(actionConfig))
}

func testOutputPlanIsUsedInApply(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, actionConfig, false, fixturePath, "run-all plan -out=plan.out --terragrunt-log-level debug")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", actionConfig.iacName)

	output = runAction(t, actionConfig, false, fixturePath, "run-all apply plan.out --terragrunt-log-level debug")
	assert.Contains(t, output, "1 added, 0 changed, 0 destroyed", actionConfig.iacName)
}

func testGitWorkingAction(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-git-commands")
	// init git repo in fixture path, run git init
	_, err := exec.Command("git", "init", fixturePath).CombinedOutput()
	if err != nil {
		t.Fatalf("Error initializing git repo: %v", err)
	}

	output := runAction(t, actionConfig, true, fixturePath, "run-all plan -out=plan.out --terragrunt-log-level debug")
	assert.Contains(t, output, fetchIacType(actionConfig)+" has been successfully initialized!", actionConfig.iacName)
	assert.Contains(t, output, "execute_INPUT_POST_EXEC_1", actionConfig.iacName)
	assert.Contains(t, output, "execute_INPUT_PRE_EXEC_1", actionConfig.iacName)
}

func testRunAllIsExecuted(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, actionConfig, false, fixturePath, "run-all plan")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", actionConfig.iacName)

	output = runAction(t, actionConfig, false, fixturePath, "run-all apply")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy", actionConfig.iacName)

	output = runAction(t, actionConfig, false, fixturePath, "run-all destroy")
	assert.Contains(t, output, "0 to add, 0 to change, 1 to destroy", actionConfig.iacName)
	assert.Contains(t, output, "Destroy complete! Resources: 1 destroyed", actionConfig.iacName)
}

func testAutoApproveDelete(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-dependencies-project")

	output := runAction(t, actionConfig, false, fixturePath, "run-all plan -out=plan.out")
	assert.Contains(t, output, "1 to add, 0 to change, 0 to destroy")

	output = runAction(t, actionConfig, false, fixturePath, "run-all apply plan.out")
	assert.Contains(t, output, "1 added, 0 changed, 0 destroyed", actionConfig.iacName)

	// run destroy with auto-approve
	output = runAction(t, actionConfig, false, fixturePath, "run-all plan -destroy -out=destroy.out")
	assert.Contains(t, output, "0 to add, 0 to change, 1 to destroy", actionConfig.iacName)

	output = runAction(t, actionConfig, false, fixturePath, "run-all apply -destroy destroy.out")
	assert.Contains(t, output, "Resources: 0 added, 0 changed, 1 destroyed", actionConfig.iacName)

	// check that fixturePath can remove recursively
	err := os.RemoveAll(fixturePath)
	assert.NoError(t, err)
}

func runAction(t *testing.T, action ActionConfig, sshAgent bool, tgPath, command string) string {
	logId := random.Random(1, 5000)
	envVars := []string{
		"INPUT_" + action.iacType + "_VERSION=" + action.iacVersion,
		"INPUT_TG_VERSION=" + action.tgVersion,
		"INPUT_TG_COMMAND=" + command,
		"INPUT_PRE_EXEC_1=echo 'execute_INPUT_PRE_EXEC_1'",
		"INPUT_POST_EXEC_1=echo 'execute_INPUT_POST_EXEC_1'",
		fmt.Sprintf("GITHUB_OUTPUT=/tmp/github-action-logs.%d", logId),
	}

	if action.tgDir != "" {
		envVars = append(envVars, "INPUT_TG_DIR="+action.tgDir)
	} else {
		envVars = append(envVars, "INPUT_TG_DIR=/github/workspace")

	}

	opts := &docker.RunOptions{
		EnvironmentVariables: envVars,
		Volumes: []string{
			tgPath + ":/github/workspace",
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
	return docker.Run(t, action.tag, opts)
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
