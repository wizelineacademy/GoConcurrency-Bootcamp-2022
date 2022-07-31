package use_cases

import (
	"GoConcurrency-Bootcamp-2022/models"
	"context"
	"fmt"
	"strings"
	"sync"
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
	var errorGlobal error
	var wg = sync.WaitGroup{}

	numWorkers := 20

	// generate channel by nums id
	chGenerator := make(chan int)
	go f.generatorChannel(from, to, chGenerator)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(chGenerator chan int) {
			defer wg.Done()
			for {
				select {
				case p, ok := <-chGenerator:
					if !ok {
						return
					}
					pokemon, err := f.api.FetchPokemon(p)

					if err != nil {
						errorGlobal = err
						// propagated error all
						cancel()
					}

					var flatAbilities []string
					for _, t := range pokemon.Abilities {
						flatAbilities = append(flatAbilities, t.Ability.URL)
					}
					pokemon.FlatAbilityURLs = strings.Join(flatAbilities, "|")

					pokemons = append(pokemons, pokemon)
				case <-ctx.Done():
					fmt.Println("error propagated")
					return
				}
			}
		}(chGenerator)
	}

	wg.Wait()

	if errorGlobal != nil {
		return errorGlobal
	}

	return f.storage.Write(pokemons)
}

// generatorChannel generate multiple channel
func (f Fetcher) generatorChannel(from, to int, chGenerator chan int) {
	for id := from; id <= to; id++ {
		chGenerator <- id
	}

	close(chGenerator)
}
