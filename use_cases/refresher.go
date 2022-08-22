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
		//done         = make(chan struct{})
		errChan = make(chan models.PokeError)
		wg      = sync.WaitGroup{}
		err     error
	)
	defer close(errChan)

	records, err := r.Read()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
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
			r.savePokemonsToCache(
				r.fetchPokemonsWithAbilities(
					parsePokemons(recordsBatch, errChan, &wg, ctx),
				),
			)
		}
	}

	go func() {
		for pokeError := range errChan {
			if pokeError.Error != nil {
				//debug
				fmt.Printf("ERROR FOUND: %s\n", pokeError.Error)
				cancel()
				//	done <- struct{}{} // signal the other pipeline stages to stop
				err = pokeError.Error
			}
		}
	}()

	wg.Wait()
	fmt.Println("FINISHED WAITING")

	if err != nil {
		fmt.Printf("Error found: %s\nAborting processing\n", err.Error())
	}

	return nil
}

func test(
	pokemonsChan chan models.Pokemon,
	//	done chan struct{},
	errChan chan models.PokeError,
	wg *sync.WaitGroup,
	ctx context.Context,
) (
	chan models.Pokemon,
	//	chan struct{},
	chan models.PokeError,
	*sync.WaitGroup,
	context.Context,
) {
	go func() {
		defer wg.Done()
		wg.Add(1)
		for {
			select {
			case <-ctx.Done():
				//debug
				fmt.Println("TEST: ERROR FOUND, ABORTING")
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
	return pokemonsChan, errChan, wg, ctx
}

func parsePokemons(recordsBatch [][]string,
	errChan chan models.PokeError,
	wg *sync.WaitGroup,
	ctx context.Context) (
	chan models.Pokemon,
	chan models.PokeError,
	*sync.WaitGroup,
	context.Context,
) {
	pokemonsChan := make(chan models.Pokemon, cons.BATCH_SIZE)

	go func() {
		//debug
		fmt.Println("PARSING POKEMON")
		defer wg.Done()
		defer close(pokemonsChan)
		wg.Add(1)
		for _, r := range recordsBatch {
			select {
			case <-ctx.Done():
				//debug
				fmt.Println("parsePokemons: ERROR FOUND, ABORTING")
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

	return pokemonsChan, errChan, wg, ctx
}

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

func (r Refresher) fetchPokemonsWithAbilities(
	pokemonsBatch chan models.Pokemon,
	errChan chan models.PokeError,
	wg *sync.WaitGroup,
	ctx context.Context,
) (
	chan models.Pokemon,
	chan models.PokeError,
	*sync.WaitGroup,
	context.Context,
) {
	outBatch := make(chan models.Pokemon, cons.BATCH_SIZE)
	go func() {
		//debug
		fmt.Println("fetchPokemonsWithAbilities")
		defer wg.Done()
		defer close(outBatch)
		wg.Add(1)
		for pokemon := range pokemonsBatch {
			select {
			case <-ctx.Done():
				//debug
				fmt.Println("fetchPokemonsWithAbilities: ERROR FOUND, ABORTING")
				return
			default:
				pokemon, err := r.fetchPokemonWithAbilities(pokemon)
				if err != nil {
					errChan <- models.PokeError{err}
					return
				}
				outBatch <- pokemon
			}
		}
	}()
	return outBatch, errChan, wg, ctx
}

func (r Refresher) savePokemonsToCache(
	inStreamBatch chan models.Pokemon,
	errChan chan models.PokeError,
	wg *sync.WaitGroup,
	ctx context.Context,
) (
	chan models.Pokemon,
	chan models.PokeError,
	*sync.WaitGroup,
	context.Context,
) {
	pokemonsBatch := []models.Pokemon{}
	go func() {
		//debug
		fmt.Println("savePokemonsToCache")
		defer wg.Done()
		defer func() {
			if len(pokemonsBatch) > 0 {
				fmt.Println("SAVING AND EXITING")
				if err := r.Save(context.Background(), pokemonsBatch); err != nil {
					fmt.Printf("ERROR WHILE SAVING: %s\n", err.Error())
					//errChan <- models.PokeError{err}
					fmt.Println("FINISHED SAVING ERR TO CHANNEL")
				}
				fmt.Println("FINISHED SAVING TO CACHE")
			}
			fmt.Println("RETURNING")
			return
		}()

		wg.Add(1)
		for pokemon := range inStreamBatch {
			select {
			case <-ctx.Done():
				//debug
				fmt.Println("savePokemonsToCache: ERROR FOUND, ABORTING")
				return
			default:
				pokemonsBatch = append(pokemonsBatch, pokemon)
			}
		}
	}()
	return inStreamBatch, errChan, wg, ctx
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
