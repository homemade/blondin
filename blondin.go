package blondin

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/homemade/blondin/alias"
)

type Balancer interface {
	Next() string
}

// An efficient solution for weighted random should be the Alias method https://en.wikipedia.org/wiki/Alias_method
// See https://www.keithschwarz.com/darts-dice-coins/ for a fuller explanation
type weightedByPercentageUsingAliasMethod struct {
	choices     []string
	aliasMethod *alias.Alias
	rng         *rand.Rand
}

func (b weightedByPercentageUsingAliasMethod) Next() string {
	i := b.aliasMethod.Gen(b.rng)
	return b.choices[i]
}

func WeightedByPercentage(config string) (Balancer, error) {

	var err error

	var choices []string
	var percentages []float64

	if config != "" { // config should be in the format <choice>:<percentage> e.g. choiceA:20,choiceB:30,choiceC:50
		configList := strings.Split(config, ",")
		for _, choice := range configList {
			choiceList := strings.Split(choice, ":")
			if len(choiceList) == 2 {
				name := choiceList[0]
				var percentage float64
				percentage, err = strconv.ParseFloat(choiceList[1], 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse percentage %s for %s %v", choiceList[1], name, err)
				}
				choices = append(choices, name)
				percentages = append(percentages, percentage)
			}

		}
	}

	// check the configured choices and percentages add up
	if len(choices) != len(percentages) {
		return nil, fmt.Errorf("invalid configuration %s, %d choices with %d percentages", config, len(choices), len(percentages))
	}
	totalPercent := 0.0
	for _, p := range percentages {
		totalPercent = totalPercent + p
	}
	if totalPercent != 100.0 {
		return nil, fmt.Errorf("invalid configuration %s, total percentages add up to %.2f", config, totalPercent)
	}

	var aliasMethod *alias.Alias
	aliasMethod, err = alias.New(percentages)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise alias method %v", err)
	}

	seed := len(choices)
	rng := rand.New(rand.NewSource(int64(seed)))

	return weightedByPercentageUsingAliasMethod{
		choices:     choices,
		aliasMethod: aliasMethod,
		rng:         rng,
	}, nil

}
