package patcher

type Runner struct{}

func NewRunner() Runner {
	return Runner{}
}

type Executor interface {
	Run() error
	CombinedOutput() ([]byte, error)
}

func (r Runner) Run(command Executor) error {
	return command.Run()
}

func (r Runner) CombinedOutput(command Executor) ([]byte, error) {
	return command.CombinedOutput()
}
