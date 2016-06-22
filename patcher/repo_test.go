package patcher_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/knit/patcher"
	"github.com/pivotal-cf-experimental/knit/patcher/fakes"
)

const gitModulesContent = `
[submodule "src/module1"]
	path = src/module-one
	url = https://example.com/module-incubator/module-one.git
[submodule "src/module2"]
	path = src/module-two
	url = https://example.com/module-incubator/module-two.git
[submodule "src/module3"]
	path = src/module-three
	url = https://example.com/module-incubator/module-three.git
[submodule "src/whoops/does/not/exist"]
	path = src/whoops/does/not/exist
	url = https://example.com/module-incubator/module-three.git
`

var _ = Describe("Repo", func() {
	var (
		runner                *fakes.CommandRunner
		repoPath              string
		repoSubmoduleFilepath string
		r                     patcher.Repo
	)

	BeforeEach(func() {
		runner = &fakes.CommandRunner{}
		var err error
		repoPath, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		var dirs = []string{"module-one", "module-two", "module-three"}
		for _, dir := range dirs {
			err := os.MkdirAll(filepath.Join(repoPath, "src", dir), 0744)
			Expect(err).NotTo(HaveOccurred())
		}

		repoSubmoduleFilepath = filepath.Join(repoPath, ".gitmodules")

		err = ioutil.WriteFile(repoSubmoduleFilepath, []byte(gitModulesContent), 0666)
		Expect(err).NotTo(HaveOccurred())
		r = patcher.NewRepo(runner, repoPath, "testbot", "foo@example.com")
	})

	AfterEach(func() {
		err := os.RemoveAll(repoPath)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("ConfigureCommitter", func() {
		It("sets the git committer name and email", func() {
			err := r.ConfigureCommitter()
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCall.Receives.Commands).To(Equal([]patcher.Command{
				patcher.Command{
					Args: []string{"config", "--global", "user.name", "testbot"},
					Dir:  repoPath,
				},
				patcher.Command{
					Args: []string{"config", "--global", "user.email", "foo@example.com"},
					Dir:  repoPath,
				},
			}))
		})

		Context("failure cases", func() {
			Context("when a config command fails", func() {
				It("returns an error", func() {
					runner.RunCall.Returns.Errors = []error{nil, errors.New("some error")}
					err := r.ConfigureCommitter()
					Expect(runner.RunCall.Count).To(Equal(2))
					Expect(err).To(MatchError("some error"))
				})
			})
		})
	})

	Describe("Checkout", func() {
		It("moves the repoistory to the specified ref", func() {
			err := r.Checkout("some-ref")
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCall.Receives.Commands).To(Equal([]patcher.Command{
				patcher.Command{
					Args: []string{"checkout", "some-ref"},
					Dir:  repoPath,
				},
				patcher.Command{
					Args: []string{"clean", "-ffd"},
					Dir:  repoPath,
				},
				patcher.Command{
					Args: []string{"submodule", "update", "--init", "--recursive", "--force"},
					Dir:  repoPath,
				},
			}))
		})

		Context("failure cases", func() {
			Context("when the checkout fails", func() {
				It("returns an error", func() {
					runner.RunCall.Returns.Errors = []error{errors.New("some error")}
					err := r.Checkout("invalid-ref")
					Expect(err).To(MatchError("some error"))
				})
			})
		})
	})

	Describe("CleanSubmodules", func() {
		It("cleans all of the submodule dirs", func() {
			err := r.CleanSubmodules()
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCall.Receives.Commands).To(Equal([]patcher.Command{
				patcher.Command{
					Args: []string{"clean", "-ffd"},
					Dir:  filepath.Join(repoPath, "src", "module-one"),
				},
				patcher.Command{
					Args: []string{"clean", "-ffd"},
					Dir:  filepath.Join(repoPath, "src", "module-two"),
				},
				patcher.Command{
					Args: []string{"clean", "-ffd"},
					Dir:  filepath.Join(repoPath, "src", "module-three"),
				},
			}))
		})

		Context("failure cases", func() {
			Context("when the .gitmodules file cannot be read", func() {
				BeforeEach(func() {
					err := os.Remove(repoSubmoduleFilepath)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					err := r.CleanSubmodules()
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when one of the clean command calls fails", func() {
				It("returns an error", func() {
					runner.RunCall.Returns.Errors = []error{nil, errors.New("some error"), nil}
					err := r.CleanSubmodules()
					Expect(runner.RunCall.Count).To(Equal(2))
					Expect(err).To(MatchError("some error"))
				})
			})
		})
	})

	Describe("ApplyPatch", func() {
		It("applies the provided top-level patches", func() {
			err := r.ApplyPatch("some-dir/something.patch")
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCall.Receives.Commands).To(Equal([]patcher.Command{
				patcher.Command{
					Args: []string{"am", "some-dir/something.patch"},
					Dir:  repoPath,
				},
			}))
		})

		Context("when an error occurs", func() {
			Context("when the command fails", func() {
				It("returns an error", func() {
					runner.RunCall.Returns.Errors = []error{errors.New("meow")}
					err := r.ApplyPatch("some-dir/something.patch")
					Expect(err).To(MatchError("meow"))
				})
			})
		})
	})

	Describe("BumpSubmodule", func() {
		It("bumps the given submodule to the provided sha", func() {
			err := r.BumpSubmodule("src/some/path", "a-sha")
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCall.Receives.Commands).To(Equal([]patcher.Command{
				patcher.Command{
					Args: []string{"checkout", "a-sha"},
					Dir:  filepath.Join(repoPath, "src", "some/path"),
				},
				patcher.Command{
					Args: []string{"submodule", "update", "--init", "--recursive", "--force"},
					Dir:  filepath.Join(repoPath, "src", "some/path"),
				},
				patcher.Command{
					Args: []string{"clean", "-ffd"},
					Dir:  filepath.Join(repoPath, "src", "some/path"),
				},
				patcher.Command{
					Args: []string{"add", "-A", "src/some/path"},
					Dir:  repoPath,
				},
				patcher.Command{
					Args: []string{"commit", "-m", "Knit bump of src/some/path", "--no-verify"},
					Dir:  repoPath,
				},
			}))
		})

		It("bumps a submodule of a submodule", func() {
			err := r.BumpSubmodule("src/some/path/src/some/other/path", "a-sha")
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCall.Receives.Commands).To(Equal([]patcher.Command{
				patcher.Command{
					Args: []string{"checkout", "a-sha"},
					Dir:  filepath.Join(repoPath, "src/some/path", "src/some/other/path"),
				},
				patcher.Command{
					Args: []string{"submodule", "update", "--init", "--recursive", "--force"},
					Dir:  filepath.Join(repoPath, "src/some/path", "src/some/other/path"),
				},
				patcher.Command{
					Args: []string{"clean", "-ffd"},
					Dir:  filepath.Join(repoPath, "src/some/path", "src/some/other/path"),
				},
				patcher.Command{
					Args: []string{"add", "-A", "src/some/other/path"},
					Dir:  filepath.Join(repoPath, "src/some/path"),
				},
				patcher.Command{
					Args: []string{"commit", "-m", "Knit bump of src/some/other/path", "--no-verify"},
					Dir:  filepath.Join(repoPath, "src/some/path"),
				},
			}))
		})

		Context("when an error occurs", func() {
			Context("when the command fails", func() {
				It("returns an error", func() {
					runner.RunCall.Returns.Errors = []error{errors.New("meow")}
					err := r.BumpSubmodule("src/some/path", "a-sha")
					Expect(err).To(MatchError("meow"))
				})
			})
		})
	})

	Describe("PatchSubmodule", func() {
		It("patches a submodule with the proper patch", func() {
			err := r.PatchSubmodule("src/different/path", "/full/submodule/some.patch")
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.CombinedOutputCall.Receives.Commands).To(Equal([]patcher.Command{
				patcher.Command{
					Args: []string{"add", "-A", "src/different/path"},
					Dir:  repoPath,
				},
			}))

			Expect(runner.RunCall.Receives.Commands).To(Equal([]patcher.Command{
				patcher.Command{
					Args: []string{"am", "/full/submodule/some.patch"},
					Dir:  filepath.Join(repoPath, "src", "different/path"),
				},
				patcher.Command{
					Args: []string{"add", "-A", "."},
					Dir:  repoPath,
				},
				patcher.Command{
					Args: []string{"commit", "-m", "Knit patch of src/different/path", "--no-verify"},
					Dir:  repoPath,
				},
			}))
		})

		Context("when the changes fail to add", func() {
			BeforeEach(func() {
				runner.CombinedOutputCall.Returns.Errors = []error{errors.New("some patch error")}
				runner.CombinedOutputCall.Returns.Outputs = [][]byte{[]byte(`fatal patchspec is in submodule 'src/some/crazy/submodule'`)}
			})

			It("adds and commits the underlying submodule", func() {
				err := r.PatchSubmodule("src/different/path", "/full/submodule/some.patch")
				Expect(err).NotTo(HaveOccurred())

				Expect(runner.CombinedOutputCall.Receives.Commands).To(Equal([]patcher.Command{
					patcher.Command{
						Args: []string{"add", "-A", "src/different/path"},
						Dir:  repoPath,
					},
				}))

				Expect(runner.RunCall.Receives.Commands).To(Equal([]patcher.Command{
					patcher.Command{
						Args: []string{"am", "/full/submodule/some.patch"},
						Dir:  filepath.Join(repoPath, "src", "different/path"),
					},
					patcher.Command{
						Args: []string{"add", "-A", "."},
						Dir:  filepath.Join(repoPath, "src/some/crazy/submodule"),
					},
					patcher.Command{
						Args: []string{"commit", "-m", "Knit submodule patch of src/some/crazy/submodule", "--no-verify"},
						Dir:  filepath.Join(repoPath, "src/some/crazy/submodule"),
					},
					patcher.Command{
						Args: []string{"add", "-A", "."},
						Dir:  repoPath,
					},
					patcher.Command{
						Args: []string{"commit", "-m", "Knit patch of src/different/path", "--no-verify"},
						Dir:  repoPath,
					},
				}))
			})
		})

		Context("when an error occurs", func() {
			Context("when the apply command fails", func() {
				It("returns an error", func() {
					runner.RunCall.Returns.Errors = []error{errors.New("meow")}
					err := r.PatchSubmodule("who-cares", "nope")
					Expect(err).To(MatchError("meow"))
				})
			})
		})
	})

	Describe("CheckoutBranch", func() {
		It("checks out the desired branch", func() {
			runner.RunCall.Returns.Errors = []error{errors.New("meow"), nil}

			branchName := "meow"
			err := r.CheckoutBranch(branchName)
			Expect(err).NotTo(HaveOccurred())

			Expect(runner.RunCall.Receives.Commands).To(Equal([]patcher.Command{
				patcher.Command{
					Args: []string{"rev-parse", "--verify", branchName},
					Dir:  repoPath,
				},
				patcher.Command{
					Args: []string{"checkout", "-b", branchName},
					Dir:  repoPath,
				},
			}))
		})

		Context("when an error occurs", func() {
			Context("when the branch already exists", func() {
				It("returns an error", func() {
					err := r.CheckoutBranch("meow")
					Expect(err).To(MatchError(`Branch "meow" already exists. Please delete it before trying again`))
					Expect(runner.RunCall.Count).To(Equal(1))
				})
			})
		})
	})
})
