# cliordle
A command-line interface version of everyone's favourite 5-letter-word-guessing game

## Installation via Go
```
$ go get github.com/j985chen/cli-ordle
```

## Usage
```
# To play a game
$ ./cliordle play

# To change gameplay settings
$ ./cliordle settings [--highContrast={true|false}] [--hardMode={true|false}]

# To view player stats
$ ./cliordle stats
```

## To-do
1. Implement hard mode
2. Clean up the cli using Cobra

## Sources
* [The original Wordle game](https://www.nytimes.com/games/wordle/index.html), for the initial inspiration & many moments of entertainment and frustration
* [cwackerfuss's react-wordle](https://github.com/cwackerfuss/react-wordle), for the wordlist & code for guess-processing
