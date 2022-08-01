package use_cases

import (
	"context"
	"strings"
	"sync"

	"GoConcurrency-Bootcamp-2022/models"
)

type reader interface {
	Read(ctx context.Context, cancel context.CancelFunc) <-chan models.Pokemon
}

// type reader interface {
// 	Read() ([]models.Pokemon, error)
// }

type saver interface {
	Save(context.Context, models.Pokemon) error
}

type fetcher interface {
	FetchAbility(string) (models.Ability, error)
}

type Refresher struct {
	reader
	saver
	fetcher
}

type AbilityResult struct {
	Error   error
	Pokemon *models.Ability
}

func NewRefresher(reader reader, saver saver, fetcher fetcher) Refresher {
	return Refresher{reader, saver, fetcher}
}

func readCSV(ctx context.Context, cancel context.CancelFunc, r Refresher) <-chan models.Pokemon {

	return r.Read(ctx, cancel)

}

func (r Refresher) Refresh(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	savePokemon(ctx, cancel, buildPokemon(ctx, readCSV(ctx, cancel, r), r), r)

	return nil
}

func buildPokemon(ctx context.Context, in <-chan models.Pokemon, r Refresher) <-chan models.Pokemon {

	wk1 := pokeWorker(ctx, in, r)
	wk2 := pokeWorker(ctx, in, r)
	wk3 := pokeWorker(ctx, in, r)

	return fanIn(wk1, wk2, wk3)

}

func savePokemon(ctx context.Context, cancel context.CancelFunc, in <-chan models.Pokemon, r Refresher) {
	for p := range in {
		if err := r.Save(ctx, p); err != nil {
			cancel()
			return
		}
	}
}

func pokeWorker(ctx context.Context, in <-chan models.Pokemon, r Refresher) <-chan models.Pokemon {

	out := make(chan models.Pokemon)
	go func() {
		defer close(out)
		for pokemon := range in {
			urls := strings.Split(pokemon.FlatAbilityURLs, "|")
			var abilities []string

			abilityChan := abilityGeneretor(ctx, r, urls)

			for ability := range abilityChan {
				abilities = append(abilities, ability)
			}
			pokemon.EffectEntries = abilities

			select {
			case <-ctx.Done():
				return
			default:
				out <- pokemon
			}

		}
	}()
	return out
}

func abilityGeneretor(ctx context.Context, r Refresher, urls []string) <-chan string {
	abilityChan := make(chan string)
	wg := sync.WaitGroup{}

	for _, url := range urls {
		wg.Add(1)
		go func(ctx context.Context, r Refresher, url string) {
			defer wg.Done()
			ability, err := r.FetchAbility(url)
			if err != nil {
				return
			}

			for _, ee := range ability.EffectEntries {
				select {
				case <-ctx.Done():
					return
				default:
					abilityChan <- ee.Effect
				}
			}
		}(ctx, r, url)

	}

	go func() {
		wg.Wait()
		close(abilityChan)
	}()

	return abilityChan

}

func fanIn(inputs ...<-chan models.Pokemon) <-chan models.Pokemon {
	var wg sync.WaitGroup
	out := make(chan models.Pokemon)
	wg.Add(len(inputs))

	for _, in := range inputs {
		go func(ch <-chan models.Pokemon) {

			for {
				value, ok := <-ch
				if !ok {
					wg.Done()
					break
				}

				out <- value
			}

		}(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
