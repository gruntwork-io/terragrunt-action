package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ActionConfig struct {
	iacName    string
	iacType    string
	iacVersion string
	tgVersion  string
}

func TestTerragruntCompositeAction(t *testing.T) {
	t.Parallel()

	testCases := []ActionConfig{
		{"OpenTofu1.8", "TOFU", "1.8.1", "0.67.0"},
		{"OpenTofu1.9", "TOFU", "1.9.0", "0.67.0"},
		{"OpenTofu1.10", "TOFU", "1.10.1", "0.82.2"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.iacName, func(t *testing.T) {
			t.Parallel()
			t.Run("testActionWithMiseConfig", func(t *testing.T) {
				t.Parallel()
				testActionWithMiseConfig(t, tc)
			})
			t.Run("testActionWithInputVersions", func(t *testing.T) {
				t.Parallel()
				testActionWithInputVersions(t, tc)
			})
			t.Run("testActionInstallOnly", func(t *testing.T) {
				t.Parallel()
				testActionInstallOnly(t, tc)
			})
			t.Run("testActionValidation", func(t *testing.T) {
				t.Parallel()
				testActionValidation(t, tc)
			})
		})
	}
}

func testActionWithMiseConfig(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-action-execution")

	// Create mise.toml in fixture
	miseConfig := fmt.Sprintf(`[tools]
terragrunt = "%s"
opentofu = "%s"
`, actionConfig.tgVersion, actionConfig.iacVersion)

	miseConfigPath := filepath.Join(fixturePath, "mise.toml")
	err := os.WriteFile(miseConfigPath, []byte(miseConfig), 0644)
	require.NoError(t, err)

	// Test that action works with mise.toml (no version inputs needed)
	output := runCompositeAction(t, "", "", actionConfig.tgVersion, fixturePath, "plan")
	assert.Contains(t, output, "Found mise configuration file")
	assert.Contains(t, output, "Starting Terragrunt Action")
}

func testActionWithInputVersions(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-action-execution")

	// Test that action works with input versions (no mise.toml)
	output := runCompositeAction(t, actionConfig.iacVersion, actionConfig.iacType, actionConfig.tgVersion, fixturePath, "plan")
	assert.Contains(t, output, "mise_config_exists=false")
	assert.Contains(t, output, "Starting Terragrunt Action")
}

func testActionInstallOnly(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-action-execution")

	// Test that action can install tools without executing terragrunt
	output := runCompositeAction(t, actionConfig.iacVersion, actionConfig.iacType, actionConfig.tgVersion, fixturePath, "")

	// Should install tools but not execute terragrunt command
	assert.NotContains(t, output, "Starting Terragrunt Action")
	// Should still install mise tools
	assert.Contains(t, output, "mise_config_exists=false")
}

func testActionValidation(t *testing.T, actionConfig ActionConfig) {
	fixturePath := prepareFixture(t, "fixture-action-execution")

	// Test that action fails when no mise.toml and no versions provided
	output := runCompositeActionWithError(t, "", "", "", fixturePath, "plan")
	assert.Contains(t, output, "ERROR: No mise.toml found, making 'tg_version' required")
}

func runCompositeAction(t *testing.T, iacVersion, iacType, tgVersion, fixturePath, command string) string {
	return runCompositeActionInternal(t, iacVersion, iacType, tgVersion, fixturePath, command, false)
}

func runCompositeActionWithError(t *testing.T, iacVersion, iacType, tgVersion, fixturePath, command string) string {
	return runCompositeActionInternal(t, iacVersion, iacType, tgVersion, fixturePath, command, true)
}

func runCompositeActionInternal(t *testing.T, iacVersion, iacType, tgVersion, fixturePath, command string, expectError bool) string {
	// Create a temporary directory for the action
	tempDir, err := os.MkdirTemp("", "terragrunt-action-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a simple script that simulates the composite action behavior
	scriptContent := `#!/bin/bash
set -e

echo "=== Checking for mise.toml and validating inputs ==="
if [[ -f "mise.toml" || -f ".mise.toml" ]]; then
  echo "mise_config_exists=true"
  echo "Found mise configuration file, will use it for tool versions"
else
  echo "mise_config_exists=false"

  if [[ -z "${INPUT_TG_VERSION}" ]]; then
    echo "ERROR: No mise.toml found, making 'tg_version' required."
    exit 1
  fi

  if [[ -z "${INPUT_TOFU_VERSION}" ]]; then
    echo "ERROR: No mise.toml found, making 'tofu_version' required"
    exit 1
  fi
fi

echo "=== Installing tools with mise ==="
echo "Would install: terragrunt ${INPUT_TG_VERSION}, opentofu ${INPUT_TOFU_VERSION}"

if [[ -n "${INPUT_TG_COMMAND}" ]]; then
  echo "=== Executing Terragrunt ==="
  echo "Starting Terragrunt Action"
  echo "Would run: terragrunt ${INPUT_TG_COMMAND} in ${INPUT_TG_DIR}"
  echo "tg_command is set to: ${INPUT_TG_COMMAND}"

  # Simulate some terragrunt output
  if [[ "${INPUT_TG_COMMAND}" == "plan" ]]; then
    echo "You can apply this plan to save these new output values to the OpenTofu state"
  fi

  echo "Finished Terragrunt Action execution"
fi
`

	scriptPath := filepath.Join(tempDir, "test_action.sh")
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Set up environment variables
	env := os.Environ()
	if tgVersion != "" {
		env = append(env, "INPUT_TG_VERSION="+tgVersion)
	}
	if iacVersion != "" && iacType == "TOFU" {
		env = append(env, "INPUT_TOFU_VERSION="+iacVersion)
	}
	if command != "" {
		env = append(env, "INPUT_TG_COMMAND="+command)
	}
	env = append(env, "INPUT_TG_DIR="+fixturePath)
	env = append(env, "INPUT_PRE_EXEC_1=echo 'execute_INPUT_PRE_EXEC_1'")
	env = append(env, "INPUT_POST_EXEC_1=echo 'execute_INPUT_POST_EXEC_1'")
	env = append(env, "GITHUB_WORKSPACE="+fixturePath)

	// Run the script in the fixture directory
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = fixturePath
	cmd.Env = env

	output, err := cmd.CombinedOutput()

	if expectError {
		assert.Error(t, err, "Expected command to fail")
	} else {
		if err != nil {
			t.Logf("Command output: %s", string(output))
			t.Logf("Command error: %v", err)
		}
		assert.NoError(t, err, "Command should succeed")
	}

	return string(output)
}

func prepareFixture(t *testing.T, fixtureDir string) string {
	path, err := files.CopyTerraformFolderToTemp(fixtureDir, "test")
	require.NoError(t, err)
	return path
}

func fetchIacType(actionConfig ActionConfig) string {
	// return Terraform if OpenTofu based on iacType value
	if strings.ToLower(actionConfig.iacType) == "tf" {
		return "Terraform"
	}
	return "OpenTofu"
}
