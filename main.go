package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// todo figure whether config should be upper or lower case
type Config struct {
	Previous string
	Next     string
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

var availableCommands map[string]cliCommand

func main() {
	config := Init()
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
		err := commandObj.callback(&config)
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

func Init() Config {
	availableCommands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "displays map and move to next location",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "displays previous location and move to it",
			callback:    commandMapb,
		},
	}
	return Config{
		Previous: "",
		Next:     "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20",
	}
}

func commandExit(config *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *Config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")

	for _, command := range availableCommands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}

	return nil
}

type locationApiResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"results"`
}

func commandMap(config *Config) error {
	var mapUrl string = config.Next
	res, err := http.Get(mapUrl)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("response failed with status code: %v", res.StatusCode)
	}

	var mapApiRes locationApiResponse

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&mapApiRes); err != nil {
		return err
	}

	config.Next = mapApiRes.Next
	config.Previous = mapApiRes.Previous

	for _, location := range mapApiRes.Results {
		fmt.Println(location.Name)
	}

	return nil
}

func commandMapb(config *Config) error {
	if config.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}
	var mapUrl string = config.Previous
	res, err := http.Get(mapUrl)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("response failed with status code: %v", res.StatusCode)
	}

	var mapApiRes locationApiResponse

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&mapApiRes); err != nil {
		return err
	}

	config.Next = mapApiRes.Next
	config.Previous = mapApiRes.Previous

	for _, location := range mapApiRes.Results {
		fmt.Println(location.Name)
	}

	return nil
}

func CleanInput(text string) []string {
	text = strings.TrimSpace(strings.ToLower(text))

	return strings.Fields(text)
}
