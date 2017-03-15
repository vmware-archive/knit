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
	Context("when everything is great", func() {
		AfterEach(func() {
			command := exec.Command("git", "checkout", "HEAD")
			command.Dir = cfReleaseRepo
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))

			command = exec.Command("git", "branch", "-D", "1.6.15")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
		})

		It("applies patches onto a clean repo", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", cfReleaseRepo,
				"-patch-repository", cfPatchesDir,
				"-version", "1.6.15")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10m").Should(gexec.Exit(0))

			Eventually(session.Out).Should(gbytes.Say("Submodule path"))

			command = exec.Command("git", "status")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("On branch 1.6.15"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("nothing to commit"))

			command = exec.Command("git", "log", "--format=%s", "-n", "8")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit bump of src/uaa"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit bump of src/etcd-release"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit bump of src/consul-release"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("add golang 1.5.3 to main blobs.yml, needed by new consul release"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/uaa"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/uaa"))
		})

		It("does not print any logs when --quiet flag is provided", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", cfReleaseRepo,
				"-patch-repository", cfPatchesDir,
				"-quiet",
				"-version", "1.6.15")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10m").Should(gexec.Exit(0))

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).NotTo(gbytes.Say(`Knit bump of src/uaa`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit bump of src/etcd-release`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit bump of src/consul-release`))
			Expect(session.Out).NotTo(gbytes.Say(`add golang 1\.5\.3 to main blobs\.yml, needed by new consul release`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit patch of src/uaa`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit patch of src/uaa`))
		})
	})

	Context("when the version specified has no starting version", func() {
		AfterEach(func() {
			command := exec.Command("git", "checkout", "HEAD")
			command.Dir = cfReleaseRepo
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))

			command = exec.Command("git", "branch", "-D", "1.6.111222")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
		})

		It("works just fine", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", cfReleaseRepo,
				"-patch-repository", cfPatchesDir,
				"-version", "1.6.111222")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m").Should(gexec.Exit(0))

			command = exec.Command("git", "status")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("On branch 1.6.111222"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("nothing to commit"))
		})
	})

	Context("when the version specified indicates a hotfix release", func() {
		AfterEach(func() {
			command := exec.Command("git", "checkout", "HEAD")
			command.Dir = cfReleaseRepo
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))

			command = exec.Command("git", "branch", "-D", "1.7.11+ipsec.uptime")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
		})

		It("applies the hotfix patches on top of the vanilla patches", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", cfReleaseRepo,
				"-patch-repository", cfPatchesDir,
				"-version", "1.7.11+ipsec.uptime")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m").Should(gexec.Exit(0))

			command = exec.Command("git", "status")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("On branch 1.7.11+ipsec.uptime"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("nothing to commit"))

			command = exec.Command("git", "log", "--format=%s", "-n", "8")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/github.com/cloudfoundry/gorouter"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/github.com/cloudfoundry/gorouter"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/github.com/cloudfoundry-incubator/route-registrar"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/capi-release/src/cloud_controller_ng"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/capi-release"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Update nginx to 1.11.1"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/loggregator"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/loggregator"))
		})
	})

	Context("when the version specified bypasses a hotfix release", func() {
		AfterEach(func() {
			command := exec.Command("git", "checkout", "HEAD")
			command.Dir = cfReleaseRepo
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))

			command = exec.Command("git", "branch", "-D", "1.7.12")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
		})

		It("does not apply the hotfix patches from the previous release", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", cfReleaseRepo,
				"-patch-repository", cfPatchesDir,
				"-version", "1.7.12")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m").Should(gexec.Exit(0))

			command = exec.Command("git", "status")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("On branch 1.7.12"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("nothing to commit"))

			command = exec.Command("git", "log", "--format=%s", "-n", "8")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/capi-release/src/cloud_controller_ng"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/capi-release/src/cloud_controller_ng"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit bump of src/consul-release"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Bump src/consul-release"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/github.com/cloudfoundry/gorouter"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/capi-release/src/cloud_controller_ng"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit patch of src/capi-release"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Update nginx to 1.11.1"))
		})
	})

	Context("when the version specified adds a new submodule", func() {
		AfterEach(func() {
			command := exec.Command("git", "checkout", "HEAD")
			command.Dir = diegoReleaseRepo
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))

			command = exec.Command("git", "branch", "-D", "1.7.15")
			command.Dir = diegoReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
		})

		It("adds the new submodule", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", diegoReleaseRepo,
				"-patch-repository", diegoPatchesDir,
				"-version", "1.7.15")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m").Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit addition of src/github.com/nats-io/nats"))
			Expect(string(session.Out.Contents())).To(ContainSubstring("Knit addition of src/github.com/nats-io/nuid"))

			gitModulesContents, err := ioutil.ReadFile(path.Join(diegoReleaseRepo, ".gitmodules"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(gitModulesContents)).To(ContainSubstring("src/github.com/nats-io/nats"))
			Expect(string(gitModulesContents)).To(ContainSubstring("src/github.com/nats-io/nuid"))

			Expect(path.Join(diegoReleaseRepo, "src/github.com/nats-io/nats")).To(BeADirectory())
			Expect(path.Join(diegoReleaseRepo, "src/github.com/nats-io/nuid")).To(BeADirectory())

			command = exec.Command("git", "rev-parse", "HEAD")
			command.Dir = path.Join(diegoReleaseRepo, "src/github.com/nats-io/nats")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("c0ad3f079763c06c3ce94ad12fa3f17e78966d99"))

			command = exec.Command("git", "rev-parse", "HEAD")
			command.Dir = path.Join(diegoReleaseRepo, "src/github.com/nats-io/nuid")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("a5152d67cf63cbfb5d992a395458722a45194715"))
		})
	})

	Context("when the version specified removes an old submodule", func() {
		AfterEach(func() {
			command := exec.Command("git", "checkout", "HEAD")
			command.Dir = cfReleaseRepo
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))

			command = exec.Command("git", "branch", "-D", "1.8.35")
			command.Dir = cfReleaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
		})

		It("removes the old submodule", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", cfReleaseRepo,
				"-patch-repository", cfPatchesDir,
				"-version", "1.8.35")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10m").Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(ContainSubstring("Submodule 'src/buildpacks' (https://github.com/cloudfoundry/buildpack-releases) unregistered for path 'src/buildpacks'"))

			gitModulesContents, err := ioutil.ReadFile(path.Join(cfReleaseRepo, ".gitmodules"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(gitModulesContents)).NotTo(ContainSubstring("src/buildpacks"))

			Expect(path.Join(cfReleaseRepo, "src/buildpacks")).NotTo(BeADirectory())
		})
	})

	Context("error cases", func() {
		Context("version branch already exists", func() {
			BeforeEach(func() {
				command := exec.Command("git", "checkout", "v222")
				command.Dir = cfReleaseRepo
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				command = exec.Command("git", "checkout", "-b", "1.6.1")
				command.Dir = cfReleaseRepo
				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
			})

			AfterEach(func() {
				command := exec.Command("git", "checkout", "HEAD")
				command.Dir = cfReleaseRepo
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				command = exec.Command("git", "branch", "-D", "1.6.1")
				command.Dir = cfReleaseRepo
				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				command = exec.Command("git", "clean", "-ffd")
				command.Dir = cfReleaseRepo
				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
			})

			It("returns an error", func() {
				command := exec.Command(patcher,
					"-repository-to-patch", cfReleaseRepo,
					"-patch-repository", cfPatchesDir,
					"-version", "1.6.1")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5m").Should(gexec.Exit(1))

				Eventually(session.Err).Should(gbytes.Say(`Branch "1.6.1" already exists. Please delete it before trying again`))
			})
		})

		Context("when flags are not set", func() {
			DescribeTable("missing flags",
				func(version, release, patch, errorString string) {
					command := exec.Command(patcher,
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
				command := exec.Command(patcher,
					"-repository-to-patch", cfReleaseRepo,
					"-patch-repository", cfPatchesDir,
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
				command := exec.Command(patcher,
					"-repository-to-patch", cfReleaseRepo,
					"-patch-repository", cfPatchesDir,
					"-version", "1.6.15")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "1m").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say("knit requires a version of git >= 2.9.0"))
			})
		})
	})
})
