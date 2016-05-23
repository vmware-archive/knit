package patcher_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pivotal-cf-experimental/knit/patcher"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandRunner", func() {
	Describe("Run", func() {
		var (
			runner patcher.CommandRunner
			stderr *bytes.Buffer
			stdout *bytes.Buffer
		)

		BeforeEach(func() {
			runner = patcher.NewCommandRunner()
			stderr = bytes.NewBuffer([]byte{})
			stdout = bytes.NewBuffer([]byte{})
		})

		It("runs the given command and returns stdout", func() {
			err := runner.Run(patcher.Command{
				Executable: "echo",
				Args: []string{
					"banana",
				},
				Stdout: stdout,
				Stderr: stderr,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(bytes.NewBuffer([]byte("banana\n"))))
			Expect(stderr).To(Equal(bytes.NewBuffer([]byte{})))
		})

		It("runs the command in the given directory", func() {
			tempDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			tempDir, err = filepath.EvalSymlinks(tempDir)
			Expect(err).NotTo(HaveOccurred())

			err = runner.Run(patcher.Command{
				Executable: "pwd",
				Dir:        tempDir,
				Stdout:     stdout,
				Stderr:     stderr,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(bytes.NewBuffer([]byte(fmt.Sprintf("%s\n", tempDir)))))
			Expect(stderr).To(Equal(bytes.NewBuffer([]byte{})))
		})

		It("includes stderr output", func() {
			err := runner.Run(patcher.Command{
				Executable: "curl",
				Args: []string{
					"-v",
					"https://google.com",
				},
				Stdout: stdout,
				Stderr: stderr,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stderr).To(ContainSubstring("GET / HTTP/1.1"))
		})

		Context("failure cases", func() {
			Context("when the given executable does not exist", func() {
				It("returns an error", func() {
					err := runner.Run(patcher.Command{
						Executable: "not-an-executable",
					})
					Expect(err).To(MatchError(ContainSubstring("executable file not found in $PATH")))
				})
			})

			Context("when the command fails to run", func() {
				It("returns an error", func() {
					err := runner.Run(patcher.Command{
						Executable: "ls",
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
		)

		BeforeEach(func() {
			runner = patcher.NewCommandRunner()
		})

		It("runs the given command and returns stdout", func() {
			output, err := runner.CombinedOutput(patcher.Command{
				Executable: "echo",
				Args: []string{
					"command output",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal([]byte("command output\n")))
		})
	})
})
