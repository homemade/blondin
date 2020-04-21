package tests

import (
	"math"
	"testing"

	"github.com/homemade/blondin"
)

func TestWeightedByPercentage(t *testing.T) {

	// configure a new WeightedByPercentage balancer
	config := "choiceA:20,choiceB:30,choiceC:50"
	balancer, err := blondin.WeightedByPercentage(config)
	if err != nil {
		t.Fatalf("failed to configure a new WeightedByPercentage balancer %v", err)
	}

	results := map[string]int{
		"choiceA": 0,
		"choiceB": 0,
		"choiceC": 0,
	}
	for i := 0; i < 100000; i++ {
		s := balancer.Next()
		results[s] = results[s] + 1
	}

	// check results
	if math.RoundToEven(float64(results["choiceA"])/1000.0) != 20 ||
		math.RoundToEven(float64(results["choiceB"])/1000.0) != 30 ||
		math.RoundToEven(float64(results["choiceC"])/1000.0) != 50 {
		t.Error("results are outside of expected distribution")
	}

	t.Logf("%#v", results)

}
