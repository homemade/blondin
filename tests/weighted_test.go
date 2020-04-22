// go test -v github.com/homemade/blondin/tests
package tests

import (
	"math"
	"testing"

	"github.com/homemade/blondin"
)

func TestWeightedByPercentage(t *testing.T) {

	for i := 5; i <= 8; i++ {

		// reset results at the start of each test
		results := map[string]int{
			"choiceA": 0,
			"choiceB": 0,
			"choiceC": 0,
		}

		// configure a new WeightedByPercentage balancer for each test
		config := "choiceA:20,choiceB:30,choiceC:50"

		balancer, err := blondin.WeightedByPercentage(config)
		if err != nil {
			t.Fatalf("failed to configure a new WeightedByPercentage balancer %v", err)
		}

		count := int(math.Pow(10, float64(i)))
		for i := 0; i < count; i++ {
			s := balancer.Next()
			results[s] = results[s] + 1
		}

		percent := func(choice string) float64 {
			return float64(results[choice]*100) / float64(count)
		}

		// log results
		t.Logf("results after %d choices", count)
		for k, v := range results {
			t.Logf("%v %v (%f)", k, v, percent(k))
		}

		// check distribution
		if math.RoundToEven(percent("choiceA")) != 20 ||
			math.RoundToEven(percent("choiceB")) != 30 ||
			math.RoundToEven(percent("choiceC")) != 50 {
			t.Errorf("results are outside of expected distribution")
		}

	}

}
