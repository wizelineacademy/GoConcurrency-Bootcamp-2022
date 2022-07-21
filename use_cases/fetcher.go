package use_cases

import (
	"log"
	"sort"
	"strings"

	"GoConcurrency-Bootcamp-2022/models"
)

type api interface {
	FetchPokemon(id int) (models.Pokemon, error)
}

type writer interface {
	Write(pokemons []models.Pokemon) error
}

type Fetcher struct {
	api     api
	storage writer
}

func NewFetcher(api api, storage writer) Fetcher {
	return Fetcher{api, storage}
}

type PokemonsByID []models.Pokemon

func (x PokemonsByID) Len() int           { return len(x) }
func (x PokemonsByID) Less(i, j int) bool { return x[i].ID < x[j].ID }
func (x PokemonsByID) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func (f Fetcher) Fetch(from, to int) error {
	var pokemons []models.Pokemon

	ch := GetResponses(from, to, f)

	for id := from; id <= to; id++ {
		pokemon := <-ch
		pokemons = append(pokemons, pokemon)
	}

	sort.Sort(PokemonsByID(pokemons))

	return f.storage.Write(pokemons)
}

func GetResponses(from, to int, f Fetcher) <-chan models.Pokemon {
	ch := make(chan models.Pokemon)
	for id := from; id <= to; id++ {
		go PingAPI(id, f, ch)
	}

	return ch
}

func PingAPI(id int, f Fetcher, ch chan models.Pokemon) {
	pokemon, err := f.api.FetchPokemon(id)
	if err != nil {
		log.Printf("ERROR: fetch pokemon failed")
	}

	var flatAbilities []string
	for _, t := range pokemon.Abilities {
		flatAbilities = append(flatAbilities, t.Ability.URL)
	}
	pokemon.FlatAbilityURLs = strings.Join(flatAbilities, "|")

	ch <- pokemon
}
