package patcher_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pivotal-cf/knit/patcher"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandRunner", func() {
	Describe("Run", func() {
		var (
			runner patcher.CommandRunner
			err    error
		)

		It("runs the given command and returns stdout", func() {
			runner, err = patcher.NewCommandRunner("echo", true)
			Expect(err).NotTo(HaveOccurred())
			runner.Stderr = bytes.NewBuffer([]byte{})
			runner.Stdout = bytes.NewBuffer([]byte{})

			err = runner.Run(patcher.Command{
				Args: []string{
					"banana",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.Stdout).To(Equal(bytes.NewBuffer([]byte("banana\n"))))
			Expect(runner.Stderr).To(Equal(bytes.NewBuffer([]byte{})))
		})

		It("runs the command in the given directory", func() {
			tempDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			tempDir, err = filepath.EvalSymlinks(tempDir)
			Expect(err).NotTo(HaveOccurred())

			runner, err = patcher.NewCommandRunner("pwd", true)
			Expect(err).NotTo(HaveOccurred())
			runner.Stderr = bytes.NewBuffer([]byte{})
			runner.Stdout = bytes.NewBuffer([]byte{})

			err = runner.Run(patcher.Command{
				Dir: tempDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.Stdout).To(Equal(bytes.NewBuffer([]byte(fmt.Sprintf("%s\n", tempDir)))))
			Expect(runner.Stderr).To(Equal(bytes.NewBuffer([]byte{})))
		})

		It("includes stderr output", func() {
			runner, err = patcher.NewCommandRunner("curl", true)
			Expect(err).NotTo(HaveOccurred())
			runner.Stderr = bytes.NewBuffer([]byte{})
			runner.Stdout = bytes.NewBuffer([]byte{})

			err := runner.Run(patcher.Command{
				Args: []string{
					"-v",
					"https://google.com",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.Stderr).To(ContainSubstring("GET / HTTP/1.1"))
		})

		Context("failure cases", func() {
			Context("when the given executable does not exist", func() {
				It("returns an error", func() {
					_, err = patcher.NewCommandRunner("not-an-executable", true)
					Expect(err).To(MatchError(ContainSubstring("executable file not found in $PATH")))
				})
			})

			Context("when the command fails to run", func() {
				It("returns an error", func() {
					runner, err = patcher.NewCommandRunner("ls", true)

					err := runner.Run(patcher.Command{
						Args: []string{
							"/some/missing/directory",
						},
					})
					Expect(err).To(MatchError(ContainSubstring("exit status")))
				})
			})
		})
	})

	Describe("CombinedOutput", func() {
		var (
			runner patcher.CommandRunner
			err    error
		)

		BeforeEach(func() {
			runner, err = patcher.NewCommandRunner("echo", true)
		})

		It("runs the given command and returns stdout", func() {
			output, err := runner.CombinedOutput(patcher.Command{
				Args: []string{
					"command output",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal([]byte("command output\n")))
		})
	})
})
