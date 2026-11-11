package sbom

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

func RunValidator(validator string, args ...string) error {
	if validator == "" {
		return nil
	}
	validatorPath := validator
	if !filepath.IsAbs(validator) {
		if _, err := os.Stat(validator); err != nil {
			if runfilePath, runfileErr := runfiles.Rlocation(validator); runfileErr == nil {
				validatorPath = runfilePath
			}
		}
	}
	cmd := exec.Command(validatorPath, args...)
	env := os.Environ()
	if os.Getenv("PATH") == "" {
		env = append(env, "PATH=/usr/bin:/bin:/usr/sbin:/sbin")
	}
	cmd.Env = env
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running validator %s: %w\n%s", validator, err, output.String())
	}
	if output.Len() > 0 {
		fmt.Fprint(os.Stderr, output.String())
	}
	return nil
}
