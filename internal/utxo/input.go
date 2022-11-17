package utxo

type Input struct {
	ID string
}

func NewInput(ID string) *Input {
	return &Input{
		ID: ID,
	}
}

func NewInputFromOutput(output *Output) *Input {
	return &Input{
		ID: output.ID,
	}
}

func NewInputsFromOutputs(outputs []*Output) []*Input {
	inputs := make([]*Input, 0)
	for _, output := range outputs {
		inputs = append(inputs, NewInputFromOutput(output))
	}
	return inputs
}
