package use_cases

import (
	"context"
	"strings"

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
	pokemons, err := r.Read()
	if err != nil {
		return err
	}

	for i, p := range pokemons {
		urls := strings.Split(p.FlatAbilityURLs, "|")
		var abilities []string
		for _, url := range urls {
			ability, err := r.FetchAbility(url)
			if err != nil {
				return err
			}

			for _, ee := range ability.EffectEntries {
				abilities = append(abilities, ee.Effect)
			}
		}

		pokemons[i].EffectEntries = abilities
	}

	if err := r.Save(ctx, pokemons); err != nil {
		return err
	}

	return nil
}
