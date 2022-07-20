package use_cases

import (
	"context"
	"log"
	"strings"
	"sync"

	"GoConcurrency-Bootcamp-2022/models"
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

	dataChan := readData(r)

	// Workers
	pokemonWorker1 := getAbilities(dataChan, r)
	pokemonWorker2 := getAbilities(dataChan, r)
	pokemonWorker3 := getAbilities(dataChan, r)

	for pokemon := range fanIn(pokemonWorker1, pokemonWorker2, pokemonWorker3) {
		// log.Printf("%v.- %v\n", pokemon.ID, pokemon.Name)
		pokemons = append(pokemons, pokemon)
	}

	if err := r.Save(ctx, pokemons); err != nil {
		return err
	}

	return nil
}

func getAbilities(pokemons <-chan models.Pokemon, r Refresher) <-chan models.Pokemon {
	out := make(chan models.Pokemon)

	go func() {
		defer close(out)
		for pokemon := range pokemons {
			urls := strings.Split(pokemon.FlatAbilityURLs, "|")
			var abilities []string
			for _, url := range urls {
				ability, err := r.FetchAbility(url)
				if err != nil {
					log.Printf("ERROR: something went wrong fetching the abilities")
				}

				for _, ee := range ability.EffectEntries {
					abilities = append(abilities, ee.Effect)
				}

				pokemon.EffectEntries = abilities
			}
			out <- pokemon
		}
	}()

	return out
}

func readData(r Refresher) chan models.Pokemon {
	out := make(chan models.Pokemon)

	go func() {
		defer close(out)

		pokemons, err := r.Read()
		if err != nil {
			log.Printf("ERROR: couldn't read the csv %q", err.Error())
		}

		for _, pokemon := range pokemons {
			out <- pokemon
		}
	}()

	return out
}

func fanIn(chans ...<-chan models.Pokemon) <-chan models.Pokemon {
	out := make(chan models.Pokemon)
	wg := &sync.WaitGroup{}

	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch <-chan models.Pokemon) {
			for pokemon := range ch {
				out <- pokemon
			}
			wg.Done()
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
