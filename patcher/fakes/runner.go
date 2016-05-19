package fakes

import "github.com/pivotal-cf-experimental/knit/patcher"

type Runner struct {
	RunCall struct {
		Count    int
		Receives struct {
			Commands []patcher.Executor
		}

		Returns struct {
			Errors []error
		}
	}

	CombinedOutputCall struct {
		Count    int
		Receives struct {
			Commands []patcher.Executor
		}

		Returns struct {
			Outputs [][]byte
			Errors  []error
		}
	}
}

func (r *Runner) Run(command patcher.Executor) error {
	r.RunCall.Receives.Commands = append(r.RunCall.Receives.Commands, command)

	var err error
	if len(r.RunCall.Returns.Errors) != 0 {
		err = r.RunCall.Returns.Errors[r.RunCall.Count]
	}

	r.RunCall.Count++
	return err
}

func (r *Runner) CombinedOutput(command patcher.Executor) ([]byte, error) {
	r.CombinedOutputCall.Receives.Commands = append(r.CombinedOutputCall.Receives.Commands, command)

	var err error
	if len(r.CombinedOutputCall.Returns.Errors) != 0 {
		err = r.CombinedOutputCall.Returns.Errors[r.CombinedOutputCall.Count]
	}

	var output []byte
	if len(r.CombinedOutputCall.Returns.Outputs) != 0 {
		output = r.CombinedOutputCall.Returns.Outputs[r.CombinedOutputCall.Count]
	}

	r.CombinedOutputCall.Count++
	return output, err
}
