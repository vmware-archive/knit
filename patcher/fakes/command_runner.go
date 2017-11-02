package fakes

import "github.com/pivotal-cf/knit/patcher"

type CommandRunner struct {
	RunCall struct {
		Count    int
		Stub     func(patcher.Command) error
		Receives struct {
			Commands []patcher.Command
		}
		Returns struct {
			Errors []error
		}
	}
	CombinedOutputCall struct {
		Count    int
		Stub     func(patcher.Command) ([]byte, error)
		Receives struct {
			Commands []patcher.Command
		}
		Returns struct {
			Outputs [][]byte
			Errors  []error
		}
	}
}

func (r *CommandRunner) Run(command patcher.Command) error {
	r.RunCall.Receives.Commands = append(r.RunCall.Receives.Commands, command)
	r.RunCall.Count = r.RunCall.Count + 1

	if r.RunCall.Stub != nil {
		return r.RunCall.Stub(command)
	}
	if len(r.RunCall.Returns.Errors) <= r.RunCall.Count-1 {
		return nil
	}
	return r.RunCall.Returns.Errors[r.RunCall.Count-1]
}

func (r *CommandRunner) CombinedOutput(command patcher.Command) ([]byte, error) {
	r.CombinedOutputCall.Receives.Commands = append(r.CombinedOutputCall.Receives.Commands, command)
	r.CombinedOutputCall.Count = r.CombinedOutputCall.Count + 1

	if r.CombinedOutputCall.Stub != nil {
		return r.CombinedOutputCall.Stub(command)
	}
	index := r.CombinedOutputCall.Count - 1
	if len(r.CombinedOutputCall.Returns.Errors) <= index {
		return []byte{}, nil
	}

	return r.CombinedOutputCall.Returns.Outputs[index], r.CombinedOutputCall.Returns.Errors[index]
}
