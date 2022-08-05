package use_cases

import (
	"GoConcurrency-Bootcamp-2022/models"
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
	var ids []int
	for id := from; id <= to; id++ {
		ids = append(ids, id)
	}
	c := generator(ids)
	out := f.makeRequests(c)
	for {
		pokemon, ok := <-out
		if !ok {
			break
		}
		pokemons = append(pokemons, pokemon)
	}
	return f.storage.Write(pokemons)
}

func generator(ids []int) <-chan int {
	out := make(chan int, len(ids))
	go func() {
		for _, n := range ids {
			out <- n
		}
		close(out)
	}()
	return out
}

func (f Fetcher) makeRequests(in <-chan int) <-chan models.Pokemon {
	var (
		n   = cap(in)
		out = make(chan models.Pokemon, n)
		wg  = sync.WaitGroup{}
	)
	wg.Add(n)
	for id := range in {
		go func(id int) {
			defer wg.Done()
			pokemon, error := f.api.FetchPokemon(id)
			if error != nil {
				return
			}
			var flatAbilities []string
			for _, t := range pokemon.Abilities {
				flatAbilities = append(flatAbilities, t.Ability.URL)
			}
			pokemon.FlatAbilityURLs = strings.Join(flatAbilities, "|")
			out <- pokemon
		}(id)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
