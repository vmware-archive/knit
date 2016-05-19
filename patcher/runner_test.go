package patcher_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/knit/patcher"
	"github.com/pivotal-cf-experimental/knit/patcher/fakes"
)

var _ = Describe("Runner", func() {
	var (
		r       patcher.Runner
		command *fakes.Command
	)

	BeforeEach(func() {
		command = &fakes.Command{}
		r = patcher.NewRunner()
	})

	Describe("Run", func() {
		It("should run the command provided", func() {
			err := r.Run(command)
			Expect(err).NotTo(HaveOccurred())

			Expect(command.RunCall.WasCalled).To(BeTrue())
		})

		Context("failure cases", func() {
			BeforeEach(func() {
				command.RunCall.Returns.Error = errors.New("some runner error")
			})

			It("returns the error", func() {
				err := r.Run(command)
				Expect(err).To(MatchError("some runner error"))

				Expect(command.RunCall.WasCalled).To(BeTrue())
			})
		})
	})

	Describe("CombinedOutput", func() {
		It("should run the command and return output", func() {
			command.CombinedOutputCall.Returns.Output = []byte("wow")
			output, err := r.CombinedOutput(command)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(Equal([]byte("wow")))
		})

		Context("failure cases", func() {
			BeforeEach(func() {
				command.CombinedOutputCall.Returns.Error = errors.New("some output error")
			})

			Context("when the command fails", func() {
				It("returns an error", func() {
					_, err := r.CombinedOutput(command)
					Expect(err).To(MatchError("some output error"))
				})
			})
		})
	})
})
