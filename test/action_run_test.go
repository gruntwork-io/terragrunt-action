package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ActionConfig struct {
	IacName    string
	IacType    string
	IacVersion string
	TgVersion  string
}

type RunActionOptions struct {
	ActionConfig ActionConfig
	Tag          string
	FixturePath  string
	Command      string
	TgDir        string
	SshAgent     bool
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
		{"OpenTofu1.8", "TOFU", "1.8.0", "0.72.9"},
		{"OpenTofu1.9", "TOFU", "1.9.0", "0.72.9"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.IacName, func(t *testing.T) {
			t.Parallel()
			fixturePath := prepareFixture(t, "fixture-action-execution")

			opt := RunActionOptions{ActionConfig: tc, Tag: tag, FixturePath: fixturePath, TgDir: "/github/workspace"}

			t.Run("testActionIsExecuted", func(t *testing.T) {
				t.Parallel()
				opt.Command = "plan"
				output := runAction(t, opt)
				assert.Contains(t, output, "You can apply this plan to save these new output values to the "+fetchIacType(tc))
			})
		})
	}
}

func TestTerragruntActionEmptyPath(t *testing.T) {
	t.Parallel()
	tag := buildActionImage(t)

	fixturePath := prepareFixture(t, "fixture-action-execution")

	config := ActionConfig{
		IacName:    "OpenTofu1.9",
		IacType:    "TOFU",
		IacVersion: "1.9.0",
		TgVersion:  "0.72.9",
	}

	opt := RunActionOptions{ActionConfig: config, Tag: tag, FixturePath: fixturePath, TgDir: ""}
	opt.Command = "plan"
	output := runAction(t, opt)
	assert.Contains(t, output, "You can apply this plan to save these new output values to the "+fetchIacType(config))
}

func runAction(t *testing.T, opts RunActionOptions) string {
	logID := random.Random(1, 5000)

	envVars := []string{
		"INPUT_" + opts.ActionConfig.IacType + "_VERSION=" + opts.ActionConfig.IacVersion,
		"INPUT_TG_VERSION=" + opts.ActionConfig.TgVersion,
		"INPUT_TG_COMMAND=" + opts.Command,
		"INPUT_PRE_EXEC_1=echo 'execute_INPUT_PRE_EXEC_1'",
		"INPUT_POST_EXEC_1=echo 'execute_INPUT_POST_EXEC_1'",
		fmt.Sprintf("GITHUB_OUTPUT=/tmp/github-action-logs.%d", logID),
	}

	if len(opts.TgDir) != 0 {
		envVars = append(envVars, "INPUT_TG_DIR="+opts.TgDir)
	}

	dockerOpts := &docker.RunOptions{
		EnvironmentVariables: envVars,
		Volumes: []string{
			opts.FixturePath + ":/github/workspace",
		},
	}

	// Configure SSH agent if needed
	if opts.SshAgent {
		homeDir, err := os.UserHomeDir()
		assert.NoError(t, err)
		sshPath := filepath.Join(homeDir, ".ssh")

		socketID := random.Random(1, 5000)
		socketPath := fmt.Sprintf("/tmp/ssh-agent.sock.%d", socketID)

		sshAgentID := docker.RunAndGetID(t, "ssh-agent:local", &docker.RunOptions{
			Detach:               true,
			Remove:               true,
			EnvironmentVariables: []string{"SSH_AUTH_SOCK=" + socketPath},
			Volumes:              []string{"/tmp:/tmp", sshPath + ":/root/keys"},
		})
		defer docker.Stop(t, []string{sshAgentID}, &docker.StopOptions{})

		dockerOpts.Volumes = append(dockerOpts.Volumes, "/tmp:/tmp")
		dockerOpts.EnvironmentVariables = append(dockerOpts.EnvironmentVariables, "SSH_AUTH_SOCK="+socketPath)
	}

	return docker.Run(t, opts.Tag, dockerOpts)
}

func prepareFixture(t *testing.T, fixtureDir string) string {
	path, err := files.CopyTerraformFolderToTemp(fixtureDir, "test")
	require.NoError(t, err)
	return path
}

func fetchIacType(actionConfig ActionConfig) string {
	if strings.ToLower(actionConfig.IacType) == "tf" {
		return "Terraform"
	}
	return "OpenTofu"
}
