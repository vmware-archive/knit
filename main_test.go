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

			command = exec.Command("git", "branch", "-D", "1.7.13")
			command.Dir = releaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
		})

		It("applies patches onto a clean repo", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", releaseRepo,
				"-patch-repository", patchesRepo,
				"-version", "1.7.13")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10m").Should(gexec.Exit(0))

			Eventually(session.Out).Should(gbytes.Say("Submodule path"))

			command = exec.Command("git", "status")
			command.Dir = releaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("On branch 1.7.13"))
			Expect(session.Out).To(gbytes.Say("nothing to commit"))

			command = exec.Command("git", "log", "--pretty=format:%s", "-n", "10")
			command.Dir = releaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say(`Knit patch of src/loggregator`))
			Expect(session.Out).To(gbytes.Say(`Knit bump of src/uaa-release`))
			Expect(session.Out).To(gbytes.Say(`Knit patch of src/capi-release/src/cloud_controller_ng`))
			Expect(session.Out).To(gbytes.Say(`Knit patch of src/capi-release/src/cloud_controller_ng`))
			Expect(session.Out).To(gbytes.Say(`Knit bump of src/consul-release`))
			Expect(session.Out).To(gbytes.Say(`Bump src/consul-release`))
			Expect(session.Out).To(gbytes.Say(`Knit patch of src/github.com/cloudfoundry/gorouter`))
			Expect(session.Out).To(gbytes.Say(`Knit patch of src/capi-release/src/cloud_controller_ng`))
			Expect(session.Out).To(gbytes.Say(`Knit patch of src/capi-release`))
			Expect(session.Out).To(gbytes.Say(`Update nginx to 1.11.1`))
		})

		It("does not print any logs when --quiet flag is provided", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", releaseRepo,
				"-patch-repository", patchesRepo,
				"-quiet",
				"-version", "1.7.13")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10m").Should(gexec.Exit(0))

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).NotTo(gbytes.Say(`Knit patch of src/loggregator`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit bump of src/uaa-release`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit patch of src/capi-release/src/cloud_controller_ng`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit patch of src/capi-release/src/cloud_controller_ng`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit bump of src/consul-release`))
			Expect(session.Out).NotTo(gbytes.Say(`Bump src/consul-release`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit patch of src/github.com/cloudfoundry/gorouter`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit patch of src/capi-release/src/cloud_controller_ng`))
			Expect(session.Out).NotTo(gbytes.Say(`Knit patch of src/capi-release`))
			Expect(session.Out).NotTo(gbytes.Say(`Update nginx to 1.11.1`))
		})
	})

	Context("when the version specified has no starting version", func() {
		AfterEach(func() {
			command := exec.Command("git", "checkout", "master")
			command.Dir = releaseRepo
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))

			command = exec.Command("git", "branch", "-D", "1.6.111222")
			command.Dir = releaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
		})

		It("works just fine", func() {
			command := exec.Command(patcher,
				"-repository-to-patch", releaseRepo,
				"-patch-repository", patchesRepo,
				"-version", "1.6.111222")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m").Should(gexec.Exit(0))

			command = exec.Command("git", "status")
			command.Dir = releaseRepo
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "30s").Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("On branch 1.6.111222"))
			Expect(session.Out).To(gbytes.Say("nothing to commit"))
		})
	})

	Context("error cases", func() {
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
					"-repository-to-patch", releaseRepo,
					"-patch-repository", patchesRepo,
					"-version", "1.6.15")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "1m").Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(`"git": executable file not found in \$PATH`))
			})
		})
	})
})
