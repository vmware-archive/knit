package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Apply Patches", func() {
	var (
		repoToPatch string
		patchesDir  string
	)

	BeforeEach(func() {
		var err error
		patchesDir, err = ioutil.TempDir("", "patch-dir")
		Expect(err).NotTo(HaveOccurred())

		err = os.Mkdir(filepath.Join(patchesDir, "1.2"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		repoToPatch, err = ioutil.TempDir("", "repo-to-patch")
		Expect(err).NotTo(HaveOccurred())

		initGitRepo(repoToPatch)

		createPatch(repoToPatch, patchesDir)
	})

	AfterEach(func() {
		os.RemoveAll(repoToPatch)
		os.RemoveAll(patchesDir)
	})

	It("applies a single patch to a clean repo", func() {
		command := exec.Command(pathToKnit,
			"-repository-to-patch", repoToPatch,
			"-patch-repository", patchesDir,
			"-version", "1.2.1")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10m").Should(gexec.Exit(0))

		Eventually(session.Out).Should(gbytes.Say("Applying: a change to the file"))

		command = exec.Command("git", "status")
		command.Dir = repoToPatch
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "30s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("On branch 1.2.1"))
		Expect(string(session.Out.Contents())).To(ContainSubstring("nothing to commit"))

		command = exec.Command("git", "log", "--format=%s", "-n", "8")
		command.Dir = repoToPatch
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("a change to the file"))
		Expect(string(session.Out.Contents())).NotTo(ContainSubstring("a hotfix patch"))
	})

	It("does not print any logs when --quiet flag is provided", func() {
		command := exec.Command(pathToKnit,
			"-repository-to-patch", repoToPatch,
			"-patch-repository", patchesDir,
			"-quiet",
			"-version", "1.2.1")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10m").Should(gexec.Exit(0))

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out).NotTo(gbytes.Say("a change to the file"))
	})

	Context("when the version specified has no starting version", func() {
		It("works just fine", func() {
			command := exec.Command(pathToKnit,
				"-repository-to-patch", repoToPatch,
				"-patch-repository", patchesDir,
				"-version", "1.2.111222")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m").Should(gexec.Exit(0))

			command = exec.Command("git", "status")
			command.Dir = repoToPatch
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("On branch 1.2.111222"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("nothing to commit"))
		})
	})

	Context("when the version specified indicates a hotfix release", func() {
		It("applies the hotfix patches on top of the vanilla patches", func() {
			command := exec.Command(pathToKnit,
				"-repository-to-patch", repoToPatch,
				"-patch-repository", patchesDir,
				"-version", "1.2.1+hot.fix")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m").Should(gexec.Exit(0))

			command = exec.Command("git", "status")
			command.Dir = repoToPatch
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("On branch 1.2.1+hot.fix"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("nothing to commit"))

			command = exec.Command("git", "log", "--format=%s", "-n", "8")
			command.Dir = repoToPatch
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("a change to the file"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("a hotfix patch"))
		})
	})

	Context("when the version specified bypasses a hotfix release", func() {
		It("does not apply the hotfix patches from the previous release", func() {
			command := exec.Command(pathToKnit,
				"-repository-to-patch", repoToPatch,
				"-patch-repository", patchesDir,
				"-version", "1.2.1")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m").Should(gexec.Exit(0))

			command = exec.Command("git", "status")
			command.Dir = repoToPatch
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("On branch 1.2.1"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("nothing to commit"))

			command = exec.Command("git", "log", "--format=%s", "-n", "8")
			command.Dir = repoToPatch
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("a change to the file"))
			Expect(string(session.Out.Contents())).ToNot(ContainSubstring("a hotfix patch"))
		})
	})

	Context("when the version specified adds a new submodule", func() {
		It("adds the new submodule", func() {
			command := exec.Command(pathToKnit,
				"-repository-to-patch", repoToPatch,
				"-patch-repository", patchesDir,
				"-version", "1.2.2")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m").Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit addition of " + filepath.Join("path", "to", "kiln")))

			gitModulesContents, err := ioutil.ReadFile(filepath.Join(repoToPatch, ".gitmodules"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(gitModulesContents)).To(ContainSubstring(filepath.Join("path", "to", "kiln")))

			Expect(path.Join(repoToPatch, filepath.Join("path", "to", "kiln"))).To(BeADirectory())

			command = exec.Command("git", "rev-parse", "HEAD")
			command.Dir = filepath.Join(repoToPatch, "path", "to", "kiln")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("7c018a3cd508e0b5541014362b353cde32d5c2a7"))
		})
	})

	Context("when the version specified removes an old submodule", func() {
		It("removes the old submodule", func() {
			command := exec.Command(pathToKnit,
				"-repository-to-patch", repoToPatch,
				"-patch-repository", patchesDir,
				"-version", "1.2.3")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10m").Should(gexec.Exit(0))

			expectedStdout := "Submodule '" + filepath.Join("path", "to", "kiln") + "' (https://github.com/pivotal-cf/kiln.git) unregistered for path '" + filepath.Join("path", "to", "kiln") + "'"
			Expect(string(session.Out.Contents())).To(ContainSubstring(expectedStdout))

			gitModulesContents, err := ioutil.ReadFile(path.Join(repoToPatch, ".gitmodules"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(gitModulesContents)).NotTo(ContainSubstring(filepath.Join("path", "to", "kiln")))

			Expect(filepath.Join(repoToPatch, "path", "to", "kiln")).NotTo(BeADirectory())
		})
	})

	Context("error cases", func() {
		Context("version branch already exists", func() {
			BeforeEach(func() {
				command := exec.Command("git", "branch", "1.2.1")
				command.Dir = repoToPatch
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
			})

			It("returns an error", func() {
				command := exec.Command(pathToKnit,
					"-repository-to-patch", repoToPatch,
					"-patch-repository", patchesDir,
					"-version", "1.2.1")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5m").Should(gexec.Exit(1))

				Eventually(session.Err).Should(gbytes.Say(`Branch "1.2.1" already exists. Please delete it before trying again`))
			})
		})

		Context("when flags are not set", func() {
			DescribeTable("missing flags",
				func(version, release, patch, errorString string) {
					command := exec.Command(pathToKnit,
						"-repository-to-patch", release,
						"-patch-repository", patch,
						"-version", version)
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say(errorString))
				},
				Entry("missing version", "", "some-repo-to-patch", "some-patch-repo", "version is a required flag"),
				Entry("missing release repo", "v1", "", "some-patch-repo", "repository-to-patch is a required flag"),
				Entry("missing patch repo", "v1", "some-repo-to-patch", "", "patch-repository is a required flag"),
			)
		})

		Context("when the git executable does not exist", func() {
			var path string

			BeforeEach(func() {
				path = os.Getenv("PATH")
				os.Setenv("PATH", "")
			})

			AfterEach(func() {
				os.Setenv("PATH", path)
			})

			It("exists with exit status 1", func() {
				command := exec.Command(pathToKnit,
					"-repository-to-patch", repoToPatch,
					"-patch-repository", patchesDir,
					"-version", "1.6.15")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "1m").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`"git": executable file not found in \$PATH`))
			})
		})

		Context("when the git executable is too old", func() {
			var (
				path     string
				fakePath string
			)

			BeforeEach(func() {
				var err error
				fakePath, err = ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				fakeGit, err := os.Create(filepath.Join(fakePath, "git"))
				Expect(err).NotTo(HaveOccurred())

				_, err = fakeGit.WriteString("#!/bin/bash\necho \"git version 2.8.0\"")
				Expect(err).NotTo(HaveOccurred())

				err = fakeGit.Chmod(0700)
				Expect(err).NotTo(HaveOccurred())

				err = fakeGit.Close()
				Expect(err).NotTo(HaveOccurred())

				path = os.Getenv("PATH")
				os.Setenv("PATH", fmt.Sprintf("%s:%s", fakePath, path))
			})

			AfterEach(func() {
				os.Setenv("PATH", path)

				err := os.RemoveAll(fakePath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("exists with exit status 1", func() {
				command := exec.Command(pathToKnit,
					"-repository-to-patch", repoToPatch,
					"-patch-repository", patchesDir,
					"-version", "1.6.15")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "1m").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say("knit requires a version of git >= 2.9.0"))
			})
		})
	})
})
