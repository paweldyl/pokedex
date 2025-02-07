package common

import (
	"pokedex/internal/pokecache"
	"sync"
)

type Config struct {
	Cache          pokecache.Cache
	Previous       string
	Next           string
	CaughtPokemons CaughtPokemons
}

type CaughtPokemons struct {
	Pokemons map[string]Pokemon
	Mu       *sync.Mutex
}

type Pokemon struct {
	Name   string
	Height int
	Weight int
	Stats  map[string]int
	Types  []string
}

func (conf Config) SavePokemon(pokemon Pokemon) {
	conf.CaughtPokemons.Mu.Lock()
	defer conf.CaughtPokemons.Mu.Unlock()
	conf.CaughtPokemons.Pokemons[pokemon.Name] = pokemon
}
