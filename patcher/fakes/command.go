package fakes

type Command struct {
	RunCall struct {
		WasCalled bool

		Returns struct {
			Error error
		}
	}

	CombinedOutputCall struct {
		Returns struct {
			Output []byte
			Error  error
		}
	}
}

func (c *Command) Run() error {
	c.RunCall.WasCalled = true
	return c.RunCall.Returns.Error
}

func (c *Command) CombinedOutput() ([]byte, error) {
	return c.CombinedOutputCall.Returns.Output, c.CombinedOutputCall.Returns.Error
}
