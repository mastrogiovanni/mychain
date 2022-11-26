package utxo

import "fmt"

type EngineHandler interface {
	// Validate if output to consume and output to insert are sound with respect to the system
	Validate(input []*Output, output []*Output) error
	// Request missing inputs. Transaction is provided to enqueue it again for evaluation once Input is retrieved
	Request(transaction *Transaction, input *Input) (*Output, error)
}

type Engine struct {
	handler EngineHandler
	storage *OutputStorage
}

func NewEngine(handler EngineHandler) *Engine {
	return &Engine{
		handler: handler,
		storage: NewOutputStorage(),
	}
}

func (engine *Engine) AddOutput(output *Output) error {
	return engine.storage.AddOutput(output)
}

func (engine *Engine) Submit(transaction *Transaction) error {

	// Find existing output from transaction's input
	outputs := make([]*Output, 0, len(transaction.Inputs))
	for _, input := range transaction.Inputs {
		output, err := engine.storage.FindOutputFromInput(input)
		if err != nil {
			output, err = engine.handler.Request(transaction, input)
			if err != nil {
				return fmt.Errorf("output not found %s", input.ID)
			}
			engine.AddOutput(output)
		}
		outputs = append(outputs, output)
	}

	// Verify transaction's outputs
	for _, output := range transaction.Outputs {
		err := output.Verify()
		if err != nil {
			return err
		}
	}

	// Validate transaction:
	// A criteria could be that sum of outputs must match sum of inputs
	err := engine.handler.Validate(outputs, transaction.Outputs)
	if err != nil {
		return err
	}

	// Consume outputs
	for _, output := range outputs {
		delete(engine.storage.Outputs, output.ID)
	}

	// Insert other Outputs
	for _, output := range transaction.Outputs {
		engine.storage.Outputs[output.ID] = output
	}

	return nil
}
