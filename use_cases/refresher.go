package use_cases

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	cons "GoConcurrency-Bootcamp-2022/constant"
	"GoConcurrency-Bootcamp-2022/models"
)

type reader interface {
	Read() ([][]string, error)
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

	var (
		recordsBatch [][]string
		done         = make(chan struct{})
		errChan      = make(chan models.PokeError)
		wg           = sync.WaitGroup{}
		err          error
	)

	records, err := r.Read()
	if err != nil {
		return err
	}

	for i, record := range records {
		if i == 0 { // header
			continue
		}

		ix := (i - 1) % cons.BATCH_SIZE
		if ix == 0 {
			recordsBatch = make([][]string, cons.BATCH_SIZE)
		}

		recordsBatch[ix] = record

		if ix == cons.BATCH_SIZE-1 {
			test(parsePokemons(recordsBatch, done, errChan, &wg))
		}
	}

	go func() {
		if pokeError := <-errChan; pokeError.Error != nil {
			done <- struct{}{} // signal the other pipeline stages to stop
			err = pokeError.Error
		}
	}()

	wg.Wait()

	return err
}

func test(pokemonsChan chan models.Pokemon,
	done chan struct{},
	errChan chan models.PokeError,
	wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		wg.Add(1)
		for {
			select {
			case <-done:
				return
			default:
				p, ok := <-pokemonsChan
				if !ok {
					return
				}
				fmt.Printf("%#v\n", p)
			}
		}
	}()
}

func parsePokemons(recordsBatch [][]string,
	done chan struct{},
	errChan chan models.PokeError,
	wg *sync.WaitGroup) (
	chan models.Pokemon,
	chan struct{},
	chan models.PokeError,
	*sync.WaitGroup) {

	pokemonsChan := make(chan models.Pokemon, cons.BATCH_SIZE)

	go func() {
		defer close(pokemonsChan)
		defer wg.Done()
		wg.Add(1)
		for _, r := range recordsBatch {
			select {
			case <-done:
				return
			default:
				pokemon, err := parsePokemon(r)
				if err != nil {
					errChan <- models.PokeError{err}
					return
				}
				pokemonsChan <- pokemon
			}
		}
	}()

	return pokemonsChan, done, errChan, wg
}

/* func recordsGenerator(records [][]string) chan [][]string {
	recordsChan := make(chan [][]string)
	go func() {
		defer close(recordsChan)
		var (
			l     = 0
			batch [][]string
		)
		for i, r := range records {
			switch {
			case i == 0:
				continue
			case l == 0:
				batch = make([][]string, cons.BATCH_SIZE)
				batch[l] = r
				l++
			case l == cons.BATCH_SIZE:
				recordsChan <- batch
				l = 0
			}
		}
		if l > 0 && l < cons.BATCH_SIZE {
			recordsChan <- batch
		}
	}()
	return recordsChan
} */

/* func (r Refresher) parsePokemons(recordsBatch [][]string,
	pokeChan chan []models.Pokemon,
	done <-chan struct{},
	errChan chan models.PokeError,
) (chan []models.Pokemon, <-chan struct{}, chan models.PokeError) {
	var (
		size      = 0
		pokeBatch = make([]models.Pokemon, cons.BATCH_SIZE)
	)

	go func() {
		for _, r := range recordsBatch {
			if r == nil {
				continue
			}

			select {
			case <-done:
				if size > 0 {
					pokeChan <- pokeBatch
				}
				return
			default:
				pokemon, err := parsePokemon(r)
				switch {
				case err != nil:
					errChan <- models.PokeError{err}
					if size > 0 {
						pokeChan <- pokeBatch
					}
					return
				case size == 0:
					pokeBatch = make([]models.Pokemon, cons.BATCH_SIZE)
					fallthrough
				case size == cons.BATCH_SIZE:
					pokeChan <- pokeBatch
					size = 0
				default:
					pokeBatch[size] = pokemon
					size++
				}
			}
		}
	}()

	return pokeChan, done, errChan
}
*/
// TODO: To avoid breaking clean architecture, expose it as public function from reader layer
func parsePokemon(record []string) (models.Pokemon, error) {
	id, err := strconv.Atoi(record[0])
	if err != nil {
		return models.Pokemon{}, err
	}

	height, err := strconv.Atoi(record[2])
	if err != nil {
		return models.Pokemon{}, err
	}

	weight, err := strconv.Atoi(record[3])
	if err != nil {
		return models.Pokemon{}, err
	}

	return models.Pokemon{
		ID:              id,
		Name:            record[1],
		Height:          height,
		Weight:          weight,
		Abilities:       nil,
		FlatAbilityURLs: record[4],
		EffectEntries:   nil,
	}, err
}

/* func (r Refresher) fetchPokemonsWithAbilities(
	sourceBatchesChannel chan []models.Pokemon,
	done <-chan struct{},
	errChan chan models.PokeError,
) (chan []models.Pokemon, <-chan struct{}, chan models.PokeError) {
	go func() {
		for batch := range sourceBatchesChannel {
			select {
			case <-done:
				return
			default:
				for p := range batch {
					pokemon, err := r.fetchPokemonWithAbilities(p)
					if err != nil {
						errChan <- models.PokeError{err}
						return
					}
					outChan <- pokemon
				}
			}
		}
		errChan <- models.PokeError{}
	}()
	return outChan
} */

func (r Refresher) savePokemonsToCache(inStream <-chan models.Pokemon,
	done chan struct{},
	errChan chan models.PokeError,
	ctx context.Context) {
	pokemonsBatch := make([]models.Pokemon, 5)
	go func() {
		for p := range inStream {
			select {
			/* 			case pokeErr := <-errChan:
			if pokeErr.Error != nil {
				if len(pokemonsBatch) > 0 {
					if err := r.Save(ctx, pokemonsBatch); err != nil {
						errChan <- models.PokeError{err}
					}
				}
				return
			} */
			case <-done:
				if len(pokemonsBatch) > 0 {
					if err := r.Save(ctx, pokemonsBatch); err != nil {
						errChan <- models.PokeError{err}
					}
				}
				return
			default:
				if len(pokemonsBatch) == 5 {
					if err := r.Save(ctx, pokemonsBatch); err != nil {
						errChan <- models.PokeError{err}
						return
					}
					pokemonsBatch = make([]models.Pokemon, 5)
				} else {
					pokemonsBatch = append(pokemonsBatch, p)
				}
			}
		}
		errChan <- models.PokeError{}
	}()
}

func (r Refresher) fetchPokemonWithAbilities(p models.Pokemon) (models.Pokemon, error) {
	urls := strings.Split(p.FlatAbilityURLs, "|")
	var abilities []string
	for _, url := range urls {
		ability, err := r.FetchAbility(url)
		if err != nil {
			return models.Pokemon{}, fmt.Errorf("fetching ability from poke endpoint: %w", err)
		}

		for _, ee := range ability.EffectEntries {
			abilities = append(abilities, ee.Effect)
		}
	}

	p.EffectEntries = abilities
	return p, nil
}
