package models

import "encoding/json"

type Pokemon struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Height    int    `json:"height"`
	Weight    int    `json:"weight"`
	Abilities []struct {
		Ability struct {
			URL string `json:"url"`
		} `json:"ability"`
	} `json:"abilities,omitempty"`

	FlatAbilityURLs string   `json:"flat_ability_ur_ls"`
	EffectEntries   []string `json:"effect_entries"`
}

func (p Pokemon) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

type Ability struct {
	ID            int `json:"id"`
	EffectEntries []struct {
		Effect string `json:"effect"`
	} `json:"effect_entries"`
}
