package repositories

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"GoConcurrency-Bootcamp-2022/models"
)

type LocalStorage struct{}

const filePath = "resources/pokemons.csv"

func (l LocalStorage) Write(pokemons []models.Pokemon) error {
	file, fErr := os.Create(filePath)

	if fErr != nil {
		return fErr
	}

	defer file.Close()

	w := csv.NewWriter(file)
	records := buildRecords(pokemons)
	if err := w.WriteAll(records); err != nil {
		return err
	}

	return nil
}

func (l LocalStorage) Read(ctx context.Context, cancel context.CancelFunc) <-chan models.Pokemon {
	file, fErr := os.Open(filePath)
	if fErr != nil {
		cancel()
		return nil
	}

	return readRecords(file, ctx, cancel)
}

func readRecords(r io.Reader, ctx context.Context, cancel context.CancelFunc) <-chan models.Pokemon {
	out := make(chan models.Pokemon)
	go func(ctx context.Context, cancel context.CancelFunc) {

		reader := csv.NewReader(r)
		_, err := reader.Read()
		if err != nil {
			cancel()
			return
		}
		for {
			record, err := reader.Read()
			if err == io.EOF {
				cancel()
				break
			}
			if err != nil {
				cancel()
				continue
			}

			pokemon, err := parseCSVData(record)
			if err != nil {
				cancel()
				continue
			}

			select {
			case <-ctx.Done():
				return
			default:
				out <- *pokemon
			}

		}
		close(out)
	}(ctx, cancel)
	return out

}

func buildRecords(pokemons []models.Pokemon) [][]string {
	headers := []string{"id", "name", "height", "weight", "flat_abilities"}
	records := [][]string{headers}
	for _, p := range pokemons {
		record := fmt.Sprintf("%d,%s,%d,%d,%s",
			p.ID,
			p.Name,
			p.Height,
			p.Weight,
			p.FlatAbilityURLs)
		records = append(records, strings.Split(record, ","))
	}

	return records
}

func parseCSVData(record []string) (*models.Pokemon, error) {

	id, err := strconv.Atoi(record[0])
	if err != nil {
		return nil, err
	}

	height, err := strconv.Atoi(record[2])
	if err != nil {
		return nil, err
	}

	weight, err := strconv.Atoi(record[3])
	if err != nil {
		return nil, err
	}

	pokemon := &models.Pokemon{
		ID:              id,
		Name:            record[1],
		Height:          height,
		Weight:          weight,
		Abilities:       nil,
		FlatAbilityURLs: record[4],
		EffectEntries:   nil,
	}

	return pokemon, nil
}
