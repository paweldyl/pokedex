package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"pokedex/internal/common"
	"pokedex/internal/pokeapi"
	"pokedex/internal/pokecache"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*common.Config, []string) error
}

var availableCommands map[string]cliCommand

func main() {
	conf := Init()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Pokedex > ")
	for scanner.Scan() {
		cleanedInput := CleanInput(scanner.Text())
		if len(cleanedInput) == 0 {
			fmt.Print("Pokedex > ")
			continue
		}
		command := cleanedInput[0]
		var args []string
		if len(cleanedInput) > 1 {
			args = cleanedInput[1:len(cleanedInput)]
		}

		commandObj, ok := availableCommands[command]
		if !ok {
			fmt.Println("Unknown command")
			fmt.Print("Pokedex > ")
			continue
		}
		err := commandObj.callback(&conf, args)
		if err != nil {
			fmt.Println(err)
			fmt.Print("Pokedex > ")
			continue
		}
		fmt.Print("Pokedex > ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "shouldn't see an error scanning a string")
	}
}

func Init() common.Config {
	cache := pokecache.NewCache(60 * time.Second)
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
		"explore": {
			name:        "explore",
			description: "displays pokemons inside passed location",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "tries to catch pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "display info about seen pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "display list of caught pokemons",
			callback:    commandPokedex,
		},
	}
	return common.Config{
		Cache:    cache,
		Previous: "",
		Next:     "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20",
		CaughtPokemons: common.CaughtPokemons{
			Pokemons: map[string]common.Pokemon{},
			Mu:       &sync.Mutex{},
		},
	}
}

func commandExit(conf *common.Config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(conf *common.Config, args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")

	for _, command := range availableCommands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}

	return nil
}

func commandMap(conf *common.Config, args []string) error {
	var mapData pokeapi.LocationRes
	err := pokeapi.GetUrlData[pokeapi.LocationRes](conf, conf.Next, &mapData)
	if err != nil {
		return err
	}
	conf.Next = mapData.Next
	conf.Previous = mapData.Previous
	printLocations(mapData)
	return nil
}

func commandMapb(conf *common.Config, args []string) error {
	if conf.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	var mapData pokeapi.LocationRes
	err := pokeapi.GetUrlData[pokeapi.LocationRes](conf, conf.Previous, &mapData)
	if err != nil {
		return err
	}
	conf.Next = mapData.Next
	conf.Previous = mapData.Previous
	printLocations(mapData)
	return nil
}

func commandExplore(conf *common.Config, args []string) error {
	if len(args) == 0 {
		return errors.New("area name required")
	}
	const baseUrl = "https://pokeapi.co/api/v2/location-area/"
	areaName := args[0]
	fmt.Println("Exploring " + areaName + "...")
	url := baseUrl + areaName

	var specificLocationData pokeapi.SpecificLocationRes
	err := pokeapi.GetUrlData[pokeapi.SpecificLocationRes](conf, url, &specificLocationData)
	if err != nil {
		return err
	}
	pokemons := specificLocationData.PokemonEncounters
	if len(pokemons) > 0 {
		fmt.Println("Found Pokemon:")
	}
	for _, pokemon := range pokemons {
		fmt.Println("- " + pokemon.Pokemon.Name)
	}
	return nil
}

func commandCatch(conf *common.Config, args []string) error {
	if len(args) == 0 {
		return errors.New("area name required")
	}
	const baseUrl = "https://pokeapi.co/api/v2/pokemon/"
	pokemonName := args[0]
	fmt.Println("Throwing a Pokeball at " + pokemonName + "...")
	url := baseUrl + pokemonName

	var pokemonData pokeapi.PokemonRes
	err := pokeapi.GetUrlData[pokeapi.PokemonRes](conf, url, &pokemonData)
	if err != nil {
		return err
	}
	pokemonPower := pokemonData.BaseExperience
	pokemonCaught := rand.Float64() < successChance(float64(pokemonPower))
	if pokemonCaught {
		fmt.Println(pokemonName + " was caught!")
		var formatedStats = map[string]int{}
		for _, stat := range pokemonData.Stats {
			formatedStats[stat.Stat.Name] = stat.BaseStat
		}
		var formatedType = []string{}
		for _, pokemonType := range pokemonData.Types {
			formatedType = append(formatedType, pokemonType.Type.Name)
		}
		pokemonSaveData := common.Pokemon{
			Name:   pokemonName,
			Height: pokemonData.Height,
			Weight: pokemonData.Weight,
			Stats:  formatedStats,
			Types:  formatedType,
		}
		conf.SavePokemon(pokemonSaveData)
	} else {
		fmt.Println(pokemonName + " escaped!")
	}
	return nil
}

func commandInspect(conf *common.Config, args []string) error {
	if len(args) == 0 {
		return errors.New("pokemon name required")
	}
	pokemonName := args[0]

	conf.CaughtPokemons.Mu.Lock()
	defer conf.CaughtPokemons.Mu.Unlock()

	pokemon, found := conf.CaughtPokemons.Pokemons[pokemonName]
	if !found {
		return errors.New("pokemon not found")
	}
	fmt.Println("Name: " + pokemon.Name)
	fmt.Println("Height: " + strconv.Itoa(pokemon.Height))
	fmt.Println("Weight: " + strconv.Itoa(pokemon.Weight))
	fmt.Println("Stats:")
	for stat, val := range pokemon.Stats {
		fmt.Println("  -" + stat + ": " + strconv.Itoa(val))
	}
	fmt.Println("Types:")
	for _, pokemonType := range pokemon.Types {
		fmt.Println("  -" + pokemonType)
	}
	return nil
}

func commandPokedex(conf *common.Config, args []string) error {
	conf.CaughtPokemons.Mu.Lock()
	defer conf.CaughtPokemons.Mu.Unlock()

	if len(conf.CaughtPokemons.Pokemons) == 0 {
		fmt.Println("No pokemons caught")
		return nil
	}

	fmt.Println("Your Pokedex:")
	for _, pokemon := range conf.CaughtPokemons.Pokemons {
		fmt.Println("  -" + pokemon.Name)
	}
	return nil
}

func successChance(powerLevel float64) float64 {
	k := 100.0
	return math.Exp(-powerLevel / k)
}

func printLocations(locationsData pokeapi.LocationRes) {
	for _, location := range locationsData.Results {
		fmt.Println(location.Name)
	}
}

func CleanInput(text string) []string {
	text = strings.TrimSpace(strings.ToLower(text))

	return strings.Fields(text)
}
