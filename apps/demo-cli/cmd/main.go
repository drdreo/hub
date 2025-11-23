package main

import (
	"flag"
	"log"
	"os"

	"demo-cli/internal/config"
	"demo-cli/internal/runner"
	"demo-cli/internal/state"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "demo-cli",
		Usage: "Run code demos with sequential steps",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List available demos",
				Action:  listDemos,
			},
			{
				Name:      "start",
				Usage:     "Start a demo",
				ArgsUsage: "<demo-name>",
				Action:    startDemo,
			},
			{
				Name:    "next",
				Aliases: []string{"n"},
				Usage:   "Execute the next step",
				Action:  executeNextStep,
			},
			{
				Name:   "reset",
				Usage:  "Reset the current demo",
				Action: resetDemo,
			},
			{
				Name:   "cleanup",
				Usage:  "Clean up demo artifacts",
				Action: cleanupDemo,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func listDemos(c *cli.Context) error {
	// Find all demo YAML files in the demos directory
	entries, err := os.ReadDir("./demos")
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		log.Println("No demos found. Create YAML files in the ./demos directory.")
		return nil
	}

	log.Println("Available demos:")
	for _, entry := range entries {
		if !entry.IsDir() && hasYamlExtension(entry.Name()) {
			demoConfig, err := config.LoadConfig("./demos/" + entry.Name())
			if err != nil {
				log.Printf("  %s (error: %v)\n", entry.Name(), err)
				continue
			}
			log.Printf("  %s - %s\n", entry.Name(), demoConfig.Description)
		}
	}
	return nil
}

func hasYamlExtension(filename string) bool {
	return len(filename) > 5 && (filename[len(filename)-5:] == ".yaml" || filename[len(filename)-4:] == ".yml")
}

func startDemo(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("Missing demo name", 1)
	}

	demoName := c.Args().Get(0)

	// Check if demo exists
	demoPath := "./demos/" + demoName
	if !hasYamlExtension(demoName) {
		demoPath += ".yaml"
	}

	if _, err := os.Stat(demoPath); os.IsNotExist(err) {
		return cli.Exit("Demo not found", 1)
	}

	demoConfig, err := config.LoadConfig(demoPath)
	if err != nil {
		return err
	}

	err = state.InitState(demoName, len(demoConfig.Steps))
	if err != nil {
		return err
	}

	// Print demo info
	log.Printf("Starting demo: %s\n", demoConfig.Name)
	log.Printf("Description: %s\n", demoConfig.Description)
	log.Println("\nSteps:")
	for i, step := range demoConfig.Steps {
		log.Printf("  %d. %s\n", i+1, step.Name)
	}
	log.Println("\nRun 'demo-cli next' to proceed with the first step")
	return nil
}

func executeNextStep(c *cli.Context) error {
	// Get current state
	currentState, err := state.GetState()
	if err != nil {
		return err
	}

	if currentState.CurrentStep >= currentState.TotalSteps {
		log.Println("Demo completed! Run 'demo-cli cleanup' to clean up artifacts.")
		return nil
	}

	// Load demo config
	demoPath := "./demos/" + currentState.DemoName
	if !hasYamlExtension(currentState.DemoName) {
		demoPath += ".yaml"
	}

	demoConfig, err := config.LoadConfig(demoPath)
	if err != nil {
		return err
	}

	// Get the current step
	stepIndex := currentState.CurrentStep
	step := demoConfig.Steps[stepIndex]

	log.Printf("Executing step %d/%d: %s\n", stepIndex+1, currentState.TotalSteps, step.Name)

	// Execute step
	err = runner.ExecuteStep(step)
	if err != nil {
		return err
	}

	// Update state
	err = state.IncrementStep()
	if err != nil {
		return err
	}

	if currentState.CurrentStep+1 >= currentState.TotalSteps {
		log.Println("Demo completed! Run 'demo-cli cleanup' to clean up artifacts.")
	} else {
		nextStep := demoConfig.Steps[stepIndex+1]
		log.Printf("Next step: %s\n", nextStep.Name)
	}
	return nil
}

func resetDemo(c *cli.Context) error {
	currentState, err := state.GetState()
	if err != nil {
		return err
	}

	// Clean up first
	if err := cleanupDemo(c); err != nil {
		return err
	}

	// Then restart with properly constructed Args
	demoCtx := cli.NewContext(c.App, flag.NewFlagSet("", flag.ContinueOnError), nil)
	if err := demoCtx.Set("", currentState.DemoName); err != nil {
		return err
	}

	return startDemo(demoCtx)
}

func cleanupDemo(c *cli.Context) error {
	currentState, err := state.GetState()
	if err != nil {
		return err
	}

	// Load demo config
	demoPath := "./demos/" + currentState.DemoName
	if !hasYamlExtension(currentState.DemoName) {
		demoPath += ".yaml"
	}

	demoConfig, err := config.LoadConfig(demoPath)
	if err != nil {
		return err
	}

	// Cleanup artifacts (this is optional per your requirements)
	for _, step := range demoConfig.Steps {
		if step.Type == "generate" && step.Target != "" {
			// Remove generated files
			os.Remove(step.Target)
			log.Printf("Removed: %s\n", step.Target)
		}
	}

	// Reset state
	state.ResetState()
	log.Println("Demo cleaned up and reset")
	return nil
}
