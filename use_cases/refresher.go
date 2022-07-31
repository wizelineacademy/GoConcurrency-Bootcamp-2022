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

	lenPokemons := len(pokemons)
	slicePokemons := int64(lenPokemons / 2)

	var chanFanIn []<-chan models.Pokemon

	ctx, cancel := context.WithCancel(context.Background())

	// verify if necessary slice pokemos
	if lenPokemons > 10 {
		chanFanIn = append(chanFanIn, r.generateFainIN(pokemons[0:slicePokemons]))
		chanFanIn = append(chanFanIn, r.generateFainIN(pokemons[slicePokemons:lenPokemons]))
	}

	chanFanINMerge := r.mergefANiN(chanFanIn...)

	// create 5 worker fanOut
	chanFanOut1 := r.generateFanOut(chanFanINMerge, cancel)
	chanFanOut2 := r.generateFanOut(chanFanINMerge, cancel)
	chanFanOut3 := r.generateFanOut(chanFanINMerge, cancel)
	chanFanOut4 := r.generateFanOut(chanFanINMerge, cancel)
	chanFanOut5 := r.generateFanOut(chanFanINMerge, cancel)

	for {
		if chanFanOut1 == nil && chanFanOut2 == nil && chanFanOut3 == nil && chanFanOut4 == nil && chanFanOut5 == nil {
			break
		}

		select {
		case p, ok := <-chanFanOut1:
			if !ok {
				chanFanOut1 = nil
			}
			r.completeAbilities(pokemons, p)
		case p, ok := <-chanFanOut2:
			if !ok {
				chanFanOut2 = nil
			}
			r.completeAbilities(pokemons, p)
		case p, ok := <-chanFanOut3:
			if !ok {
				chanFanOut3 = nil
			}
			r.completeAbilities(pokemons, p)
		case p, ok := <-chanFanOut4:
			if !ok {
				chanFanOut4 = nil
			}
			r.completeAbilities(pokemons, p)
		case p, ok := <-chanFanOut5:
			if !ok {
				chanFanOut5 = nil
			}
			r.completeAbilities(pokemons, p)
		case <-ctx.Done():
			chanFanOut1 = nil
			chanFanOut2 = nil
			chanFanOut3 = nil
			chanFanOut4 = nil
			chanFanOut5 = nil
		}
	}

	// if call function cancel --> cancel context
	return r.Save(ctx, pokemons)
}

func (r Refresher) generateFainIN(pokemons []models.Pokemon) chan models.Pokemon {
	chanFanIN := make(chan models.Pokemon)
	go func() {
		for _, pokemon := range pokemons {
			chanFanIN <- pokemon
		}
		close(chanFanIN)
	}()

	return chanFanIN
}

func (r Refresher) mergefANiN(chansPokemons ...<-chan models.Pokemon) chan models.Pokemon {
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

func (r Refresher) generateFanOut(pokemons chan models.Pokemon, cancel context.CancelFunc) chan models.Pokemon {
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

func (r Refresher) completeAbilities(pokemons []models.Pokemon, pokemon models.Pokemon) {
	for i, p := range pokemons {
		if p.ID == pokemon.ID {
			pokemons[i] = pokemon
		}
	}
}
