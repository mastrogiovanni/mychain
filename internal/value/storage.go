package value

import "fmt"

func NewValueStore() *ValueStore {
	return &ValueStore{
		ledger: make(map[string]int64),
	}
}

func (vs *ValueStore) Add(output *ValueOutput) {
	owner := output.owner.String()
	current, ok := vs.ledger[owner]
	if !ok {
		vs.ledger[owner] = output.value
	} else {
		vs.ledger[owner] = current + output.value
	}
}

func (vs *ValueStore) Remove(output *ValueOutput) error {
	owner := output.owner.String()
	current, ok := vs.ledger[owner]
	if !ok {
		return fmt.Errorf("negative value")
	}
	if current < output.value {
		return fmt.Errorf("negative value")
	}
	vs.ledger[owner] = current - output.value
}
