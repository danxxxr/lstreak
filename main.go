package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// streakProb returns P(at least one losing streak of length >= L in N trades).
//
// Recurrence (DP):
//
//	P[i] = 0                                       for i < L
//	P[L] = q^L
//	P[n] = P[n-1] + q^L * p * (1 - P[n-L-1])     for n > L
//
// A streak of length L ends at position n when:
//   - trades [n-L+1 .. n] are all losses  (probability q^L)
//   - trade  [n-L]         is a win        (probability p)
//   - no streak occurred in the first [n-L-1] trades (probability 1 - P[n-L-1])
func streakProb(n, L int, winRate float64) float64 {
	if n <= 0 || L <= 0 || winRate <= 0 || winRate >= 1 {
		return 0
	}
	q := 1.0 - winRate
	p := winRate
	qL := math.Pow(q, float64(L))

	if n < L {
		return 0
	}

	dp := make([]float64, n+1)
	dp[L] = qL

	for i := L + 1; i <= n; i++ {
		prev := 0.0
		if idx := i - L - 1; idx >= 0 {
			prev = dp[idx]
		}
		dp[i] = dp[i-1] + qL*p*(1-prev)
		if dp[i] > 1.0 {
			dp[i] = 1.0
		}
	}

	return dp[n]
}

func defaultTrades() []int {
	return []int{
		10, 20, 30, 40, 50, 60, 70, 80, 90, 100,
		200, 300, 400, 500, 600, 700, 800, 900, 1000, 5000, 10000,
	}
}

func parseTrades(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil || v <= 0 {
			return nil, fmt.Errorf("invalid trade count: %q", p)
		}
		result = append(result, v)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("empty trades list")
	}
	return result, nil
}

// autoNoColor returns true when ANSI color should be disabled:
// on Windows (cmd.exe / PowerShell without VT mode) or when stdout is not a TTY.
func autoNoColor() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return true
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

// colorOf returns ANSI escape codes for the given probability value.
func colorOf(p float64) (pre, post string) {
	pct := p * 100
	switch {
	case pct < 0.005:
		return "\033[90m", "\033[0m" // grey
	case pct <= 5:
		return "\033[32m", "\033[0m" // green
	case pct <= 15:
		return "\033[92m", "\033[0m" // light green
	case pct <= 35:
		return "\033[33m", "\033[0m" // yellow
	case pct <= 55:
		return "\033[93m", "\033[0m" // light yellow
	case pct <= 75:
		return "\033[91m", "\033[0m" // light red
	default:
		return "\033[31m", "\033[0m" // red
	}
}

func printTable(winRate float64, minL, maxL int, trades []int, noColor bool) {
	fmt.Printf("\nLOSING STREAK PROBABILITIES  [win rate: %.2f%%]\n\n",
		winRate*100)

	fmt.Printf("%-10s", "trades\\L")
	for L := minL; L <= maxL; L++ {
		fmt.Printf(" %8d", L)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 10+(maxL-minL+1)*9))

	for _, n := range trades {
		fmt.Printf("%-10d", n)
		for L := minL; L <= maxL; L++ {
			p := streakProb(n, L, winRate)
			s := fmt.Sprintf("%.2f%%", p*100)
			if noColor {
				fmt.Printf(" %8s", s)
			} else {
				pre, post := colorOf(p)
				fmt.Printf(" %s%8s%s", pre, s, post)
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func main() {
	winRate := flag.Float64("wr", 50.0, "win rate in percent (e.g. 55.5)")
	minL := flag.Int("min-streak", 2, "minimum streak length column")
	maxL := flag.Int("max-streak", 15, "maximum streak length column")
	tradesStr := flag.String("trades", "", "comma-separated trade counts (default preset)")
	noColor := flag.Bool("no-color", false, "disable ANSI color output")
	single := flag.Bool("single", false, "print a single value; requires -n and -l")
	nVal := flag.Int("n", 0, "number of trades (single mode)")
	lVal := flag.Int("l", 0, "streak length (single mode)")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `lstreak — losing streak probability calculator

Usage: lstreak [options]

Options:`)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Examples:
  lstreak                           # full table, win rate 50%
  lstreak -wr 55                    # win rate 55%
  lstreak -wr 45 -max-streak 10     # streak lengths 2..10
  lstreak -trades 50,100,500,1000   # custom row set
  lstreak -single -n 100 -l 7       # single value`)
	}
	flag.Parse()

	wr := *winRate / 100.0
	if wr <= 0 || wr >= 1 {
		fmt.Fprintln(os.Stderr, "error: win rate must be between 0 and 100 (exclusive)")
		os.Exit(1)
	}
	if *minL < 1 || *maxL < *minL {
		fmt.Fprintln(os.Stderr, "error: min-streak must be >= 1 and <= max-streak")
		os.Exit(1)
	}

	if !*noColor {
		*noColor = autoNoColor()
	}

	if *single {
		if *nVal <= 0 || *lVal <= 0 {
			fmt.Fprintln(os.Stderr, "error: -single requires -n > 0 and -l > 0")
			os.Exit(1)
		}
		p := streakProb(*nVal, *lVal, wr)
		fmt.Printf("P(losing streak >= %d in %d trades | win rate %.2f%%) = %.4f%%\n",
			*lVal, *nVal, *winRate, p*100)
		return
	}

	trades := defaultTrades()
	if *tradesStr != "" {
		var err error
		trades, err = parseTrades(*tradesStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}

	printTable(wr, *minL, *maxL, trades, *noColor)
}
