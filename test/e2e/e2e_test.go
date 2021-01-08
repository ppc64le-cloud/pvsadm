package e2e

import (
	"testing"

	_ "github.com/ppc64le-cloud/pvsadm/test/e2e/qcow2ova"
)

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}
