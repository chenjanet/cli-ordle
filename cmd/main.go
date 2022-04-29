package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/j985chen/cli-ordle/words"
)

const colourGreen = "\033[42m %s \033[0m"
const colourYellow = "\033[43m %s \033[0m"

var db *bolt.DB

type Player struct {
	Settings      *Settings  `json:"settings"`
	Played        float64    `json:"played"`
	Won           float64    `json:"won"`
	CurrStreak    float64    `json:"currStreak"`
	LongestStreak float64    `json:"longestStreak"`
	Stats         [6]float64 `json:"stats"`
}

type Settings struct {
	ColourBlind bool `json:"colourBlind"`
	HardMode    bool `json:"hardMode"`
}

func (p *Player) CreateGame() error {
	answer, err := words.RandomWord()
	if err != nil {
		return err
	}
	currGame := Game{p, []string{}, answer, false}
	err = currGame.PlayGame()
	return err
}

func (p *Player) ManageSettings() error {
	var err error
	return err
}

func (p *Player) ViewStats() error {
	var winPercent float64
	if p.Played == 0 {
		winPercent = 0
	} else {
		winPercent = (p.Won / p.Played) * 100
	}
	fmt.Println("--- STATISTICS ---")
	fmt.Printf("Played: %.0f | Win%%: %.0f%% | Current streak: %.0f | Longest streak: %.0f\n", p.Played, winPercent, p.CurrStreak, p.LongestStreak)
	fmt.Println()
	fmt.Println("--- GUESS DISTRIBUTION ---")
	for i := 0; i < 6; i++ {
		fmt.Printf("%d\t|\t%f\n", i+1, p.Stats[i])
	}
	return nil
}

func (p *Player) SaveStats() error {
	playerBytes, err := json.Marshal(*p)
	if err != nil {
		return fmt.Errorf("could not marshal player data json: %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		err = tx.Bucket([]byte("DB")).Put([]byte("PLAYER"), playerBytes)
		if err != nil {
			return fmt.Errorf("could not set player data: %v", err)
		}
		return nil
	})
	return err
}

type Game struct {
	Player       *Player
	WordsGuessed []string
	Answer       string
	Solved       bool
}

func (g *Game) ProcessGuess(guess string) error {
	isValid := words.IsValidGuess(guess)
	if !isValid {
		return fmt.Errorf("invalid")
	}
	g.WordsGuessed = append(g.WordsGuessed, guess)
	if guess == g.Answer {
		g.Solved = true
	}
	return nil
}

func (g *Game) PrintBoard() error {
	fmt.Printf(" ___  ___  ___  ___  ___\n")
	for i := 0; i < len(g.WordsGuessed); i++ {
		for j := 0; j < 5; j++ {
			letter := string(g.WordsGuessed[i][j])
			actual := string(g.Answer[j])

			fmt.Printf("|")
			if letter == actual {
				fmt.Printf(string(colourGreen), letter)
			} else if strings.Contains(g.Answer, letter) {
				fmt.Printf(string(colourYellow), letter)
			} else {
				fmt.Printf(" %s ", letter)
			}
			fmt.Printf("|")
		}
		fmt.Println("\n ---  ---  ---  ---  ---")
	}
	for i := len(g.WordsGuessed); i < 6; i++ {
		for j := 0; j < 5; j++ {
			fmt.Printf("|   |")
		}
		fmt.Println("\n ---  ---  ---  ---  ---")
	}
	fmt.Println()
	return nil
}

func (g *Game) HandleResults() error {
	if g.Solved {
		fmt.Println("Impressive!")
		g.Player.CurrStreak++
		g.Player.LongestStreak = math.Max(g.Player.CurrStreak, g.Player.LongestStreak)
		g.Player.Stats[len(g.WordsGuessed)-1]++
		g.Player.Won++
	} else {
		fmt.Printf("The answer was %s\n", g.Answer)
		g.Player.CurrStreak = 0
	}
	g.Player.Played++
	err := g.Player.SaveStats()
	return err
}

func (g *Game) PlayGame() error {
	var err error
	var input string
	reader := bufio.NewReader(os.Stdin)
	for i := 1; i <= 6; i++ {
		wordErr := fmt.Errorf("invalid")
		for wordErr != nil {
			fmt.Printf("Guess %d/6: ", i)
			input, err = reader.ReadString('\n')
			guess := strings.TrimSuffix(input, "\n")
			wordErr = g.ProcessGuess(guess)
			if wordErr != nil {
				fmt.Printf("%s is an invalid guess, try again\n", guess)
			}
		}
		g.PrintBoard()
		if g.Solved {
			break
		}
	}
	g.HandleResults()
	return err
}

func setupDB() error {
	var dbErr error
	db, dbErr = bolt.Open("cliordle.db", 0600, nil)

	if dbErr != nil {
		return fmt.Errorf("could not open db, %v", dbErr)
	}

	dbErr = db.Update(func(tx *bolt.Tx) error {
		_, bucketErr := tx.CreateBucketIfNotExists([]byte("DB"))
		if bucketErr != nil {
			return fmt.Errorf("could not create root bucket: %v", bucketErr)
		}
		return nil
	})
	if dbErr != nil {
		return fmt.Errorf("could not set up buckets, %v", dbErr)
	}
	return nil
}

func initPlayer() (Player, error) {
	var player Player
	err := db.View(func(tx *bolt.Tx) error {
		playerBytes := tx.Bucket([]byte("DB")).Get([]byte("PLAYER"))
		var dbErr error = nil
		if playerBytes != nil {
			dbErr = json.Unmarshal(playerBytes, &player)
		} else {
			playerSettings := Settings{false, false}
			player = Player{&playerSettings, 0, 0, 0, 0, [6]float64{0}}
		}
		return dbErr
	})
	return player, err
}

func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func manageCommands(player *Player) error {
	// validate that correct number of arguments is being received
	if len(os.Args) < 1 {
		return errors.New("Insufficient number of arguments")
	}

	action := flag.String("action", "play", "Action to perform")

	flag.Parse()

	if !(*action == "play" || *action == "settings" || *action == "stats") {
		return errors.New("Only 'play', 'settings', and 'stats' actions are supported")
	}

	if (*action == "play" || *action == "settings" || *action == "stats") && len(os.Args) > 2 {
		return errors.New("Too many arguments specified")
	}

	var err error

	if *action == "play" {
		err = player.CreateGame()
	} else if *action == "settings" {
		err = player.ManageSettings()
	} else {
		err = player.ViewStats()
	}
	return err
}

func main() {
	dbErr := setupDB()

	if dbErr != nil {
		exitGracefully(dbErr)
	}

	defer db.Close()

	// display usage info when user enters --help option
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] \nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	player, err := initPlayer()

	// processing user command
	err = manageCommands(&player)

	if err != nil {
		exitGracefully(err)
	}
}
