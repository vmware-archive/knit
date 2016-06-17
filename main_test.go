package main_test

import (
	"os"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Apply Patches", func() {
	Context("when everything is great", func() {
		AfterEach(func() {
			command := exec.Command("git", "checkout", "master")
			command.Dir = releaseRepo
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))

			command = exec.Command("git", "branch", "-D", "1.6.15")
			command.Dir = releaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
		})

		It("applies patches onto a clean repo", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", releaseRepo,
				"-patch-repository", patchesRepo,
				"-debug",
				"-version", "1.6.15")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10m").Should(gexec.Exit(0))

			Eventually(session.Out).Should(gbytes.Say("Submodule path"))

			command = exec.Command("git", "status")
			command.Dir = releaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("On branch 1.6.15"))
			Expect(session.Out).To(gbytes.Say("nothing to commit, working directory clean"))

			command = exec.Command("git", "log", "--pretty=format:%s", "-n", "8")
			command.Dir = releaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say(`Knit bump of src/uaa`))
			Expect(session.Out).To(gbytes.Say(`Knit bump of src/etcd-release`))
			Expect(session.Out).To(gbytes.Say(`Knit bump of src/consul-release`))
			Expect(session.Out).To(gbytes.Say(`add golang 1\.5\.3 to main blobs\.yml, needed by new consul release.*`))
			Expect(session.Out).To(gbytes.Say(`Knit patch of src/uaa`))
			Expect(session.Out).To(gbytes.Say(`Knit patch of src/uaa`))
		})
	})

	Context("error cases", func() {
		Context("when the version specified has no starting version", func() {
			It("returns an error", func() {
				command := exec.Command(patcher,
					"-repository-to-patch", releaseRepo,
					"-patch-repository", patchesRepo,
					"-debug",
					"-version", "1.6.111222")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "5m").Should(gexec.Exit(1))

				Eventually(session.Err).Should(gbytes.Say(`Missing starting version "1.6.111222" in starting-versions.yml`))
			})
		})

		Context("version branch already exists", func() {
			BeforeEach(func() {
				command := exec.Command("git", "checkout", "v222")
				command.Dir = releaseRepo
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				command = exec.Command("git", "checkout", "-b", "1.6.1")
				command.Dir = releaseRepo
				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
			})

			AfterEach(func() {
				command := exec.Command("git", "checkout", "master")
				command.Dir = releaseRepo
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				command = exec.Command("git", "branch", "-D", "1.6.1")
				command.Dir = releaseRepo
				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				command = exec.Command("git", "clean", "-ffd")
				command.Dir = releaseRepo
				session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))
			})

			It("returns an error", func() {
				command := exec.Command(patcher,
					"-repository-to-patch", releaseRepo,
					"-patch-repository", patchesRepo,
					"-debug",
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
			BeforeEach(func() {
				os.Setenv("PATH", "")
			})

			It("exists with exit status 1", func() {
				command := exec.Command(patcher,
					"-repository-to-patch", releaseRepo,
					"-patch-repository", patchesRepo,
					"-debug",
					"-version", "1.6.15")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "1m").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`"git": executable file not found in \$PATH`))
			})
		})
	})
})
