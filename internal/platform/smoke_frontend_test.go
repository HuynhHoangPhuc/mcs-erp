//go:build frontend

package platform_test

import (
	"os/exec"
	"testing"
)

func TestFrontend_Builds(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("pnpm"); err != nil {
		t.Skip("pnpm not found in PATH")
	}

	cmd := exec.Command("pnpm", "build")
	cmd.Dir = "../../web"
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pnpm build failed: %v\n%s", err, string(out))
	}
}
