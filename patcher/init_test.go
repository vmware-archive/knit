package patcher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Patcher Suite")
}
