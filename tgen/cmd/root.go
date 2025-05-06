package cmd

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
)

var (
	targetURL string
	rate      int
	duration  int
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "tgen",
	Short: "Generates traffic to a server",
}

// Execute launches the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// --- Traffic Sending Logic ---

func sendTraffic(target string, rate, duration int) {
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	continuous := (duration == -1)
	var timeout <-chan time.Time
	if !continuous {
		timeout = time.After(time.Second * time.Duration(duration))
	}

	for {
		select {
		case <-ticker.C:
			go func() {
				_, err := http.Get(target)
				if err != nil {
					fmt.Println("Request failed:", err)
				}
			}()
		case <-stop:
			fmt.Println("Traffic generation stopped by user.")
			return
		case <-timeout:
			fmt.Println("Traffic test finished.")
			return
		}
	}
}

func sendRandomTraffic(target string, baseRate, duration int) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	continuous := (duration == -1)
	var timeout <-chan time.Time
	if !continuous {
		timeout = time.After(time.Second * time.Duration(duration))
	}

TrafficLoop:
	for {
		// Random spike logic
		rate := baseRate
		if rand.Float32() < 0.3 {
			rate = baseRate * (rand.Intn(10) + 2)
			fmt.Println(" Traffic spike! New rate:", rate)
		} else {
			fmt.Println(" Normal traffic. Rate:", rate)
		}

		ticker := time.NewTicker(time.Second / time.Duration(rate))
		spikeDuration := time.After(time.Duration(rand.Intn(10)+5) * time.Second)

		for {
			select {
			case <-ticker.C:
				go func() {
					_, err := http.Get(target)
					if err != nil {
						fmt.Println("Request failed:", err)
					}
				}()
			case <-stop:
				ticker.Stop()
				fmt.Println("Traffic generation stopped by user.")
				return
			case <-spikeDuration:
				ticker.Stop()
				continue TrafficLoop
			case <-timeout:
				ticker.Stop()
				fmt.Println("Traffic test finished.")
				return
			}
		}
	}
}

// --- Command Definitions ---

var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Generate consistent traffic to the target URL",
	PreRunE: validateFlags,
	Run: func(cmd *cobra.Command, args []string) {
		if duration == -1 {
			fmt.Printf("Sending continuous traffic to %s at %d for %d seconds, Use Ctrl+C to stop.\n", targetURL, rate, duration)
		} else {
			fmt.Printf("Sending traffic to %s at %d rps for %d seconds\n", targetURL, rate, duration)
		}
		sendTraffic(targetURL, rate, duration)
	},
}

var runRandomCmd = &cobra.Command{
	Use:     "run-random",
	Short:   "Generate traffic with random spikes",
	PreRunE: validateFlags,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(" Sending spiky traffic to %s at base %d rps for %d seconds\n", targetURL, rate, duration)
		sendRandomTraffic(targetURL, rate, duration)
	},
}

func validateFlags(cmd *cobra.Command, args []string) error {
	if rate <= 0 {
		return fmt.Errorf("rate must be greater than 0")
	}
	if duration < -1 {
		return fmt.Errorf("duration must be greater than 0 or -1 for continuous mode")
	}
	return nil
}

// --- CLI Initialization ---

func init() {
	// Shared flags
	for _, c := range []*cobra.Command{runCmd, runRandomCmd} {
		c.Flags().StringVarP(&targetURL, "url", "u", "", "Target URL to send requests to")
		c.Flags().IntVarP(&rate, "rate", "r", 1, "Requests per second")
		c.Flags().IntVarP(&duration, "duration", "d", 10, "Test duration in seconds (-1 for continuous)")
		c.MarkFlagRequired("url")
		rootCmd.AddCommand(c)
	}
}
