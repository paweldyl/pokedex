package main

import(
	"fmt"
	"strings"
	"bufio"
	"os"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}
		
var availableCommands map[string]cliCommand

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0);
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")

	for _, command := range availableCommands {
		fmt.Printf("%s: %s\n",command.name, command.description)
	}

	return nil
}

func main() {
	Init()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Pokedex > ")
	for scanner.Scan() {
		cleanedInput := CleanInput(scanner.Text())
		if len(cleanedInput) == 0 {
			continue
		}
		command := cleanedInput[0]
		
		commandObj, ok := availableCommands[command]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		err := commandObj.callback()
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Print("Pokedex > ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "shouldn't see an error scanning a string")
	}
}

func Init() {
	availableCommands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name: "help",
			description: "Displays a help message",
			callback: commandHelp,
		},
	}
}

func CleanInput(text string) []string {
	text = strings.TrimSpace(strings.ToLower(text))
	
	return strings.Fields(text)
}