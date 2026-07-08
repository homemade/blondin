// go test -v github.com/homemade/blondin/tests
package tests

import (
	"math"
	"strings"
	"sync"
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

// TestSeedVariesPerBalancer verifies that two freshly configured balancers do
// not replay the identical opening sequence. With the old len(choices) seed
// every balancer produced the same draws, starving minority choices at startup.
func TestSeedVariesPerBalancer(t *testing.T) {

	config := "choiceA:50,choiceB:50"

	sequence := func() string {
		balancer, err := blondin.WeightedByPercentage(config)
		if err != nil {
			t.Fatalf("failed to configure a new WeightedByPercentage balancer %v", err)
		}
		var b strings.Builder
		for range 64 {
			b.WriteString(balancer.Next())
			b.WriteByte(',')
		}
		return b.String()
	}

	// Compare several independently configured balancers. If seeding is
	// deterministic they would all be identical; with per-process seeding at
	// least one of a handful of 64-draw openings should differ.
	first := sequence()
	allIdentical := true
	for range 8 {
		if sequence() != first {
			allIdentical = false
			break
		}
	}
	if allIdentical {
		t.Errorf("all balancers produced the identical opening sequence; seed is not varying per instance")
	}

}

// TestSharedBalancerUnderConcurrency models the common consumer pattern where a
// balancer is configured once, cached, and then served to many concurrent
// callers (e.g. one per in-flight request). It guards two properties:
//   - concurrent Next() calls on the shared balancer are race-free, which the
//     race detector (go test -race) verifies via the RNG access inside Next();
//   - the aggregate distribution still matches the configured weights.
func TestSharedBalancerUnderConcurrency(t *testing.T) {

	var cache sync.Map
	config := "choiceA:20,choiceB:30,choiceC:50"

	// getBalancer mirrors a load-or-create cache: WeightedByPercentage runs at
	// most once per config, and every caller shares the resulting balancer.
	getBalancer := func() blondin.Balancer {
		if v, ok := cache.Load(config); ok {
			return v.(blondin.Balancer)
		}
		b, err := blondin.WeightedByPercentage(config)
		if err != nil {
			t.Errorf("failed to configure a new WeightedByPercentage balancer %v", err)
			return nil
		}
		actual, _ := cache.LoadOrStore(config, b)
		return actual.(blondin.Balancer)
	}

	const goroutines = 16
	const perGoroutine = 50000

	var mu sync.Mutex
	results := map[string]int{}

	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := map[string]int{}
			for range perGoroutine {
				b := getBalancer()
				if b == nil {
					return
				}
				local[b.Next()]++
			}
			mu.Lock()
			for k, v := range local {
				results[k] += v
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	total := goroutines * perGoroutine
	percent := func(choice string) float64 {
		return float64(results[choice]*100) / float64(total)
	}

	t.Logf("results after %d concurrent choices", total)
	for k, v := range results {
		t.Logf("%v %v (%f)", k, v, percent(k))
	}

	if math.RoundToEven(percent("choiceA")) != 20 ||
		math.RoundToEven(percent("choiceB")) != 30 ||
		math.RoundToEven(percent("choiceC")) != 50 {
		t.Errorf("results are outside of expected distribution")
	}

}
