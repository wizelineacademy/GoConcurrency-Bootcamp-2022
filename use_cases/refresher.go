package use_cases

import (
	"GoConcurrency-Bootcamp-2022/models"
	"context"
	"fmt"
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

	done := make(chan struct{})
	defer close(done)

	in := r.gen(done, pokemons)
	c1 := r.sq(done, in)
	c2 := r.sq(done, in)
	c3 := r.sq(done, in)

	var updatedPokemons []models.Pokemon
	for p := range r.merge(done, c1, c2, c3) {
		fmt.Println("ng-updatedp", p)
		updatedPokemons = append(updatedPokemons, p)
	}

	if err := r.Save(ctx, updatedPokemons); err != nil {
		return err
	}

	return nil
}

func (r Refresher) gen(done <-chan struct{}, pokemons []models.Pokemon) <-chan models.Pokemon {
	out := make(chan models.Pokemon)
	go func() {
		defer close(out)
		for _, p := range pokemons {
			select {
			case out <- p:
			case <-done:
				fmt.Print("gen --done")
				return
			}
		}
	}()
	return out
}
func (r Refresher) sq(done chan struct{}, in <-chan models.Pokemon) <-chan models.Pokemon {
	out := make(chan models.Pokemon)
	go func() {
		defer close(out)
		for pokemon := range in {
			urls := strings.Split(pokemon.FlatAbilityURLs, "|")
			var abilities []string
			for _, url := range urls {
				ability, err := r.FetchAbility(url)
				if err != nil {
					done <- struct{}{}
				}

				for _, ee := range ability.EffectEntries {
					abilities = append(abilities, ee.Effect)
				}
			}
			pokemon.EffectEntries = abilities

			select {
			case out <- pokemon:
			case <-done:
				fmt.Print("sq-- done")
				return
			}
		}
	}()
	return out
}
func (r Refresher) merge(done <-chan struct{}, cs ...<-chan models.Pokemon) <-chan models.Pokemon {
	var wg sync.WaitGroup
	out := make(chan models.Pokemon)
	output := func(c <-chan models.Pokemon) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
