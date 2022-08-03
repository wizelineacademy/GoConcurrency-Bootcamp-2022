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
	var pokemons []models.Pokemon
	var numWorkers = 3
	var workers []<-chan models.Pokemon

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pokemons, err := r.Read()

	if err != nil {
		return err
	}

	dataChan := r.ReadDataChannel(pokemons)

	for i := 0; i < numWorkers; i++ {
		workers = append(workers, r.getAbilities(dataChan, cancel))
	}

	for p := range r.FanIN(workers...) {
		pokemons = append(pokemons, p)
	}

	// if call function cancel --> cancel context
	return r.Save(ctx, pokemons)
}

func (r Refresher) ReadDataChannel(pokemons []models.Pokemon) chan models.Pokemon {
	ch := make(chan models.Pokemon)
	go func() {
		for _, pokemon := range pokemons {
			ch <- pokemon
		}
		close(ch)
	}()

	return ch
}

func (r Refresher) FanIN(chansPokemons ...<-chan models.Pokemon) chan models.Pokemon {
	var wg = new(sync.WaitGroup)
	out := make(chan models.Pokemon)

	send := func(chanPokemons <-chan models.Pokemon, wg *sync.WaitGroup) {
		defer wg.Done()
		for n := range chanPokemons {
			out <- n
		}
	}

	wg.Add(len(chansPokemons))

	for _, chanPokemon := range chansPokemons {
		go send(chanPokemon, wg)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (r Refresher) getAbilities(pokemons chan models.Pokemon, cancel context.CancelFunc) chan models.Pokemon {
	ch := make(chan models.Pokemon)
	go func() {
		defer close(ch)
		for p := range pokemons {
			urls := strings.Split(p.FlatAbilityURLs, "|")
			var abilities []string
			for _, url := range urls {
				ability, err := r.FetchAbility(url)

				if err != nil {
					cancel()
				}

				for _, ee := range ability.EffectEntries {
					abilities = append(abilities, ee.Effect)
				}
			}
			p.EffectEntries = abilities

			ch <- p
		}
	}()
	return ch
}
