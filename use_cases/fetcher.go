package use_cases

import (
	"strings"
	"sync"

	"GoConcurrency-Bootcamp-2022/models"
)

type api interface {
	FetchPokemon(id int, wg *sync.WaitGroup) (models.Pokemon, error)
}

type writer interface {
	Write(pokemons []models.Pokemon) error
}

type Fetcher struct {
	api     api
	storage writer
	wg      sync.WaitGroup
}

func NewFetcher(api api, storage writer) Fetcher {
	return Fetcher{api, storage, sync.WaitGroup{}}
}

func (f Fetcher) Fetch(from, to int) <-chan error {
	errChan := make(chan error)
	var pokemons []models.Pokemon
	nPokemons := to - from + 1
	f.wg.Add(nPokemons)
	for id := from; id <= to; id++ {
		go func(id int) {
			pokemon, err := f.api.FetchPokemon(id, &f.wg)
			if err != nil {
				errChan <- err
			} else {
				var flatAbilities []string
				for _, t := range pokemon.Abilities {
					flatAbilities = append(flatAbilities, t.Ability.URL)
				}
				pokemon.FlatAbilityURLs = strings.Join(flatAbilities, "|")

				pokemons = append(pokemons, pokemon)
			}

		}(id)
	}
	f.wg.Wait()
	close(errChan)
	f.storage.Write(pokemons)
	return errChan
}
