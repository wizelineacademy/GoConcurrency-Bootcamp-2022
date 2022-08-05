package use_cases

import (
	"GoConcurrency-Bootcamp-2022/models"
	"context"
	"strings"
	"sync"
)

type reader interface {
	Read() ([]models.Pokemon, error)
}

type saver interface {
	Save(context.Context, []models.Pokemon) error
}

type fetcher interface {
	FetchAbility(string) (models.Ability, error)
}

type Refresher struct {
	reader
	saver
	fetcher
}

func NewRefresher(reader reader, saver saver, fetcher fetcher) Refresher {
	return Refresher{reader, saver, fetcher}
}

func (r Refresher) Refresh(ctx context.Context) error {
	pokemons, err := r.Read()
	if err != nil {
		return err
	}

	stopSignal := make(chan struct{})
	defer close(stopSignal)

	pokemonChannel := r.generatePokemonChannel(stopSignal, pokemons)
	c1 := r.setPokemonAbility(stopSignal, pokemonChannel)
	c2 := r.setPokemonAbility(stopSignal, pokemonChannel)
	c3 := r.setPokemonAbility(stopSignal, pokemonChannel)

	var updatedPokemons []models.Pokemon
	for pokemon := range r.merge(stopSignal, c1, c2, c3) {
		updatedPokemons = append(updatedPokemons, pokemon)
	}

	if err := r.Save(ctx, updatedPokemons); err != nil {
		return err
	}

	return nil
}

func (r Refresher) generatePokemonChannel(stopSignal <-chan struct{}, pokemons []models.Pokemon) <-chan models.Pokemon {
	out := make(chan models.Pokemon)
	go func() {
		defer close(out)
		for _, p := range pokemons {
			select {
			case out <- p:
			case <-stopSignal:
				return
			}
		}
	}()
	return out
}
func (r Refresher) setPokemonAbility(stopSignal chan struct{}, in <-chan models.Pokemon) <-chan models.Pokemon {
	out := make(chan models.Pokemon)
	go func() {
		defer close(out)
		for pokemon := range in {
			urls := strings.Split(pokemon.FlatAbilityURLs, "|")
			var abilities []string
			for _, url := range urls {
				ability, err := r.FetchAbility(url)
				if err != nil {
					stopSignal <- struct{}{}
				}

				for _, ee := range ability.EffectEntries {
					abilities = append(abilities, ee.Effect)
				}
			}
			pokemon.EffectEntries = abilities

			select {
			case out <- pokemon:
			case <-stopSignal:
				return
			}
		}
	}()
	return out
}
func (r Refresher) merge(stopSignal <-chan struct{}, channels ...<-chan models.Pokemon) <-chan models.Pokemon {
	var wg sync.WaitGroup
	outChannel := make(chan models.Pokemon)
	output := func(c <-chan models.Pokemon) {
		defer wg.Done()
		for pokemon := range c {
			select {
			case outChannel <- pokemon:
			case <-stopSignal:
				return
			}
		}
	}
	wg.Add(len(channels))
	for _, c := range channels {
		go output(c)
	}
	go func() {
		wg.Wait()
		close(outChannel)
	}()
	return outChannel
}
