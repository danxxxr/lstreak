# lstreak

A command-line tool that calculates the probability of encountering at least one losing streak of a given length within a series of trades.

## Installation

```bash
go install github.com/danxxxr/lstreak@latest
```

## Usage

```
lstreak [options]

Options:
  -wr float         win rate in percent (default 50)
  -min-streak int   minimum streak length column (default 2)
  -max-streak int   maximum streak length column (default 15)
  -trades string    comma-separated trade counts (default preset)
  -no-color         disable ANSI color output
  -single           print a single value; requires -n and -l
  -n int            number of trades (single mode)
  -l int            streak length (single mode)
  -version          print version and exit
```

## Examples

Full table with default settings (win rate 50%):
```
lstreak
```

Custom win rate:
```
lstreak -wr 55
```

Limit streak columns:
```
lstreak -wr 45 -max-streak 10
```

Custom row set:
```
lstreak -trades 50,100,500,1000
```

Single value:
```
lstreak -single -n 100 -l 7
```

## How it works

For each combination of trade count N and streak length L, lstreak computes:

**P(at least one losing streak of length ≥ L in N trades)**

using the following recurrence:

```
P[i] = 0                                     for i < L
P[L] = q^L
P[n] = P[n-1] + q^L × p × (1 - P[n-L-1])   for n > L
```

where `q = 1 - win_rate` (loss rate) and `p = win_rate`.

## License

MIT