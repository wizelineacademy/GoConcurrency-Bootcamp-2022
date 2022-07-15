package use_cases

import (
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

func (f Fetcher) Fetch(from, to int) error {
	var pokemons []models.Pokemon
	for id := from; id <= to; id++ {
		pokemon, err := f.api.FetchPokemon(id)
		if err != nil {
			return err
		}

		var flatAbilities []string
		for _, t := range pokemon.Abilities {
			flatAbilities = append(flatAbilities, t.Ability.URL)
		}
		pokemon.FlatAbilityURLs = strings.Join(flatAbilities, "|")

		pokemons = append(pokemons, pokemon)
	}

	return f.storage.Write(pokemons)
}
