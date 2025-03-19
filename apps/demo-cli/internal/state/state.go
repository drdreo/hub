package state

import (
	"encoding/json"
	"os"
)

// DemoState tracks the current state of a demo
type DemoState struct {
	DemoName    string `json:"demoName"`
	CurrentStep int    `json:"currentStep"`
	TotalSteps  int    `json:"totalSteps"`
}

const stateFile = ".demo-state.json"

// InitState initializes a new demo state
func InitState(demoName string, totalSteps int) error {
	state := DemoState{
		DemoName:    demoName,
		CurrentStep: 0,
		TotalSteps:  totalSteps,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}

// GetState retrieves the current demo state
func GetState() (*DemoState, error) {
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	var state DemoState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// IncrementStep moves to the next demo step
func IncrementStep() error {
	state, err := GetState()
	if err != nil {
		return err
	}

	state.CurrentStep++

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}

// ResetState clears the demo state
func ResetState() error {
	if _, err := os.Stat(stateFile); !os.IsNotExist(err) {
		return os.Remove(stateFile)
	}
	return nil
}
