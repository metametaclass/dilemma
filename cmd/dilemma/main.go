package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// blue button -> if >= than 50% agents select blue button, everyone lives
// red button -> agent selected blue button lives, but selected blue dies if blue selected by less then <50% agents

// agent strategies:
//  select blue every time
//  select red every time
//  select at random, where threshold is separate variable in range [0..1]

// first and second variant actually reduced to third in with a=0, and a=1, so we need something like
// probability distribution for agents strategy

// we need iterative(?) solution where groups with different stragegy distribution competes in number of agents alive after decision

// winning groups mates and reproduces combined distribution?

// distribution: f(x) -> y where x - percent of agent population, y -> probability of selection of red button [0..1]
// N - number of agents alive
// maybe we just need a random probability for each agent and random inherited probability?
// we assume hermaphrodite agents and random mating?
// and assume stable population - each two mated agents produces two offsprings

// single simulation step
// returns red choices count and alive agent count
func iteration(r *rand.Rand, agents []float64, choices []bool, count int) (int, int) {
	// make a choice using strategy
	redCount := 0

	for i := 0; i < count; i++ {
		strategy := agents[i]
		isRed := r.Float64() < strategy
		choices[i] = isRed
		if isRed {
			redCount++
		}
	}
	if redCount < count/2 {
		// blue wins
		//fmt.Println("blue wins")

	} else {
		// red wins
		// fmt.Println("red wins")
		for i := 0; i < count; {
			if choices[i] {
				// agent selected red, let them live
				i++
			} else {
				// agent selected blue, death awaits them
				// swap agents and choices with last element of agents array
				agents[i] = agents[count-1]
				choices[i] = choices[count-1]
				count--
			}
		}
	}
	// shuffle alive agents
	singles := make([]int, count)
	for i := 0; i < count; i++ {
		singles[i] = i
	}
	r.Shuffle(count, func(i, j int) {
		singles[i], singles[j] = singles[j], singles[i]
	})
	// mate agents in random order
	for i := 0; i < count-1; {
		s1 := agents[singles[i]]
		s2 := agents[singles[i+1]]
		offspring1 := s1 + r.Float64()*(s2-s1)
		offspring2 := s1 + r.Float64()*(s2-s1)
		agents[singles[i]] = offspring1
		agents[singles[i+1]] = offspring2
		i += 2
	}
	// last single agent without mate lives for next iteration
	return redCount, count
}

// runDilemma main command execution, calculate simulated iterative red/blue button dilemma solution
func runDilemma(cmd *cobra.Command, args []string) error {
	agentCount, err := cmd.Flags().GetInt("agents")
	if err != nil {
		return err
	}
	iterations, err := cmd.Flags().GetInt("iterations")
	if err != nil {
		return err
	}

	slopeStart, err := cmd.Flags().GetInt("slope_start")
	if err != nil {
		return err
	}

	slopeEnd, err := cmd.Flags().GetInt("slope_end")
	if err != nil {
		return err
	}

	slopeStartF := float64(slopeStart) / 100.0
	slopeEndF := float64(slopeEnd) / 100.0
	fmt.Printf("agents:%d, iterations:%d, slope_start:%f slope_end:%f\n", agentCount, iterations, slopeStartF, slopeEndF)

	if slopeEndF <= slopeStartF {
		return fmt.Errorf("invalid strategy slope range: %f..%f", slopeStartF, slopeEndF)
	}

	agents := make([]float64, agentCount)
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	for i := range agents {
		v := r.Float64()
		if v <= slopeStartF {
			agents[i] = 0
		} else if v >= slopeEndF {
			agents[i] = 1
		} else {
			agents[i] = (v - slopeStartF) / (slopeEndF - slopeStartF)
		}

	}
	currentCount := agentCount
	choices := make([]bool, agentCount)
	for i := 0; i < iterations; i++ {
		var redCount int
		redCount, currentCount = iteration(r, agents, choices, currentCount)
		fmt.Printf("%d,%d\n", redCount, currentCount)
		if currentCount < 2 {
			break
		}
		// TODO: dump stragegy distribution histogram
	}
	return nil
}

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "dilemma [-N nnnnn]",
		RunE:          runDilemma,
		Example:       "dilemma",
		SilenceUsage:  true, // do not show usage on error
		SilenceErrors: true, // do not show errors
		Args:          cobra.ArbitraryArgs,
	}
	rootCmd.PersistentFlags().IntP("agents", "N", 150, "Initial number of agents")
	rootCmd.PersistentFlags().IntP("iterations", "I", 100, "Number of iterations")
	rootCmd.PersistentFlags().IntP("slope_start", "S", 0, "Piecewise slope start, %")
	rootCmd.PersistentFlags().IntP("slope_end", "E", 100, "Piecewise slope end, %")

	return rootCmd
}

func main() {
	err := NewRootCommand().Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		os.Exit(1)
	}
}
