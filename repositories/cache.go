package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"GoConcurrency-Bootcamp-2022/models"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	redis *redis.Client
}

const cacheKey = "pokemons"

func NewCache() Cache {
	var options = &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
	}

	return Cache{redis.NewClient(options)}
}

func (c Cache) Save(ctx context.Context, pokemon models.Pokemon) error {
	hashMap := make(map[string]interface{})

	hashMap[fmt.Sprintf("%d", pokemon.ID)] = pokemon
	return c.redis.HSet(ctx, cacheKey, hashMap).Err()
}

func (c Cache) GetPokemons(ctx context.Context) ([]models.Pokemon, error) {
	rawData, err := c.redis.HGetAll(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var result []models.Pokemon
	for _, data := range rawData {
		p := models.Pokemon{}
		if uErr := json.Unmarshal([]byte(data), &p); uErr != nil {
			return nil, uErr
		}

		result = append(result, p)
	}

	return result, nil
}
