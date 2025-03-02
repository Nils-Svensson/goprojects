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

var rootCmd = &cobra.Command{
	Use:   "tgen",
	Short: "Generates traffic to a server",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	targeturl string
	rate      int
	duration  int
)

func sendTraffic(targeturl string, rate int, duration int) {
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	timeout := time.After(time.Second * time.Duration(duration))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	for {
		select {
		case <-ticker.C: //Case executes every time the channel emits a tick
			go func() { // Goroutine to send the request
				_, err := http.Get(targeturl)
				if err != nil {
					fmt.Println("Request failed", err)
				}
			}()

		case <-stop:
			fmt.Println("Traffic generation stopped by user.")
			ticker.Stop()
			return

		case <-timeout:
			fmt.Println("Traffic test finished.")
			ticker.Stop()
			return
		}
	}
}
func sendRandomTraffic(targeturl string, baseRate int, duration int) {

	var ticker *time.Ticker
	var rate int
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt) // Catch Ctrl+C for manual stop

	continuousMode := (duration == -1)

	var timeout <-chan time.Time
	if !continuousMode {
		timeout = time.After(time.Second * time.Duration(duration)) // Only set timeout for fixed mode
	}

TrafficLoop:
	for {
		// Decide if we enter a spike mode
		if rand.Float32() < 0.3 { // 30% chance of a spike
			rate = baseRate * (rand.Intn(10) + 2) // 2x to 11x spike
			fmt.Println("Traffic spike! New rate:", rate)
		} else {
			rate = baseRate
			fmt.Println("Normal traffic. Rate:", rate)
		}

		// Update ticker to new rate
		if ticker != nil {
			ticker.Stop()
		}
		ticker = time.NewTicker(time.Second / time.Duration(rate))

		// Keep this rate for a random duration (5-15 seconds)
		spikeDuration := time.After(time.Second * time.Duration(rand.Intn(10)+5))

		for {
			select {
			case <-ticker.C:
				go func() {
					_, err := http.Get(targeturl)
					if err != nil {
						fmt.Println("Request failed:", err)
					}
				}()

			case <-stop:
				fmt.Println("Traffic generation stopped by user.")
				ticker.Stop()
				return
			case <-spikeDuration:
				continue TrafficLoop // Exit inner loop, restart outer loop
			case <-timeout:
				ticker.Stop()
				fmt.Println("Traffic test finished.")
				return
			}
		}
	}
}

var runCmd = &cobra.Command{
	Use:   "run", // command to run the function
	Short: "Generates traffic to the target url",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if rate <= 0 {

			fmt.Println("Rate must be an integer greater than 0")
			os.Exit(1)
		}
		if duration < -1 {
			fmt.Println("Duration must be an integer greater than 0 or -1 for continuous mode")
			os.Exit(1)
		}
	},

	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("Starting traffic to %s at a rate of %d requests per second for %d seconds\n", targeturl, rate, duration)
		sendTraffic(targeturl, rate, duration)

	},
}
var runRandomCmd = &cobra.Command{
	Use:   "run-random",
	Short: "Generates traffic with random spikes",
	PreRun: func(cmd *cobra.Command, args []string) {
		if rate <= 0 {

			fmt.Println("Rate must be an integer greater than 0")
			os.Exit(1)
		}
		if duration < -1 {
			fmt.Println("Duration must be an integer greater than 0 or -1 for continuous mode")
			os.Exit(1)
		}

	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Random traffic started to %s at a base rate of %d requests per second for %d seconds\n", targeturl, rate, duration)
		sendRandomTraffic(targeturl, rate, duration)

	},
}

func init() {
	runCmd.Flags().StringVarP(&targeturl, "url", "u", "", "target url")                       // flag to specify the target url
	runCmd.Flags().IntVarP(&rate, "rate", "r", 1, "rate of requests per second")              //  flag to specify the rate of requests per second, default is 1
	runCmd.Flags().IntVarP(&duration, "duration", "d", 10, "duration of the test in seconds") // flag to specify the duration of the test, default is 10 seconds
	runCmd.MarkFlagRequired("url")

	runRandomCmd.Flags().StringVarP(&targeturl, "url", "u", "", "target url")
	runRandomCmd.Flags().IntVarP(&rate, "rate", "r", 2, "rate of requests per second")
	runRandomCmd.Flags().IntVarP(&duration, "duration", "d", 10, "duration of the test in seconds")
	runRandomCmd.MarkFlagRequired("url")

	rootCmd.AddCommand(runRandomCmd) // mark the url flag as required
	rootCmd.AddCommand(runCmd)

}
