package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type TerraformResult struct {
	Success bool
	Output  string
	Error   string
}

type TerraformExecutor struct {
	workDir string
}

func NewTerraformExecutor(workDir string) *TerraformExecutor {
	return &TerraformExecutor{workDir: workDir}
}

func (e *TerraformExecutor) Prepare(workDir, appName, imageTag string, envVars map[string]string) error {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}

	tfvars := fmt.Sprintf(`app_name = "%s"
image_tag = "%s"
internal_port = 3000
host_port = 8080
`, appName, imageTag)

	for k, v := range envVars {
		tfvars += fmt.Sprintf("%s = \"%s\"\n", k, v)
	}

	if err := os.WriteFile(filepath.Join(workDir, "terraform.tfvars"), []byte(tfvars), 0644); err != nil {
		return fmt.Errorf("write tfvars: %w", err)
	}

	return nil
}

func (e *TerraformExecutor) RunInit(workDir string) TerraformResult {
	return e.runCmd(workDir, "init", "-input=false")
}

func (e *TerraformExecutor) RunPlan(workDir string) TerraformResult {
	return e.runCmd(workDir, "plan", "-input=false", "-out=tfplan")
}

func (e *TerraformExecutor) RunApply(workDir string) TerraformResult {
	return e.runCmd(workDir, "apply", "-input=false", "-auto-approve", "tfplan")
}

func (e *TerraformExecutor) RunDestroy(workDir string) TerraformResult {
	return e.runCmd(workDir, "destroy", "-input=false", "-auto-approve")
}

func (e *TerraformExecutor) runCmd(workDir, subcommand string, args ...string) TerraformResult {
	cmdArgs := append([]string{subcommand}, args...)
	cmd := exec.Command("terraform", cmdArgs...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Info().Str("dir", workDir).Strs("args", cmdArgs).Msg("executing terraform")

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		stdoutStr := strings.TrimSpace(stdout.String())
		output := stdoutStr
		if stderrStr != "" {
			output = stdoutStr + "\n" + stderrStr
		}
		return TerraformResult{
			Success: false,
			Output:  output,
			Error:   fmt.Sprintf("terraform %s failed: %s", subcommand, err.Error()),
		}
	}

	return TerraformResult{
		Success: true,
		Output:  strings.TrimSpace(stdout.String()),
	}
}
