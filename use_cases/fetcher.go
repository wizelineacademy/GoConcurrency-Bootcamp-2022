package use_cases

import (
	"context"
	"strings"
	"sync"

	"GoConcurrency-Bootcamp-2022/models"
)

type api interface {
	FetchPokemon(id int) (models.Pokemon, error)
}

type writer interface {
	Write(pokemons []models.Pokemon) error
}

type Fetcher struct {
	ctx     context.Context
	api     api
	storage writer
}

type PokeResult struct {
	Error   error
	Pokemon *models.Pokemon
}

func NewFetcher(ctx context.Context, api api, storage writer) Fetcher {
	return Fetcher{ctx, api, storage}
}

func (f Fetcher) Fetch(from, to int) error {
	var pokemons []models.Pokemon
	ctx, cancel := context.WithCancel(f.ctx)

	pokeChannel := pokemonGeneretor(ctx, f, from, to)

	for pokeResult := range pokeChannel {
		if pokeResult.Error != nil && pokeResult.Pokemon == nil {
			cancel()
			return pokeResult.Error
		}
		pokemons = append(pokemons, *pokeResult.Pokemon)
	}

	return f.storage.Write(pokemons)
}

func pokemonGeneretor(ctx context.Context, f Fetcher, from, to int) <-chan PokeResult {
	pokemonChan := make(chan PokeResult)
	wg := sync.WaitGroup{}

	for id := from; id <= to; id++ {
		wg.Add(1)
		go func(ctx context.Context, f Fetcher, id int) {
			defer wg.Done()
			var result PokeResult
			pokemon, err := f.api.FetchPokemon(id)
			result = PokeResult{Pokemon: &pokemon, Error: err}
			if err != nil {
				result.Pokemon = nil

			}

			if result.Pokemon != nil {
				var flatAbilities []string
				for _, t := range pokemon.Abilities {
					flatAbilities = append(flatAbilities, t.Ability.URL)
				}
				result.Pokemon.FlatAbilityURLs = strings.Join(flatAbilities, "|")
				result.Error = nil
			}

			select {
			case <-ctx.Done():
				return
			case pokemonChan <- result:
			}
		}(ctx, f, id)

	}

	go func() {
		wg.Wait()
		close(pokemonChan)
	}()

	return pokemonChan

}
