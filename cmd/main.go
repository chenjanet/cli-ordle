package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/j985chen/cli-ordle/words"
)

const colourGreen = "\033[42m %s \033[0m"
const colourYellow = "\033[43m %s \033[0m"

type Game struct {
	player       *Player
	wordsGuessed []string
	answer       string
	solved       bool
}

func (g *Game) ProcessGuess(guess string) error {
	isValid := words.IsValidGuess(guess)
	if !isValid {
		return fmt.Errorf("invalid")
	}
	g.wordsGuessed = append(g.wordsGuessed, guess)
	if guess == g.answer {
		g.solved = true
	}
	return nil
}

func (g *Game) PrintBoard() error {
	fmt.Printf(" ___  ___  ___  ___  ___\n")
	for i := 0; i < len(g.wordsGuessed); i++ {
		for j := 0; j < 5; j++ {
			letter := string(g.wordsGuessed[i][j])
			actual := string(g.answer[j])

			fmt.Printf("|")
			if letter == actual {
				fmt.Printf(string(colourGreen), letter)
			} else if strings.Contains(g.answer, letter) {
				fmt.Printf(string(colourYellow), letter)
			} else {
				fmt.Printf(" %s ", letter)
			}
			fmt.Printf("|")
		}
		fmt.Println("\n ---  ---  ---  ---  ---")
	}
	for i := len(g.wordsGuessed); i < 6; i++ {
		for j := 0; j < 5; j++ {
			fmt.Printf("|   |")
		}
		fmt.Println("\n ---  ---  ---  ---  ---")
	}
	fmt.Println()
	return nil
}

func (g *Game) HandleResults() error {
	var err error
	return err
}

func (g *Game) Play() error {
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
		if g.solved {
			break
		}
	}
	g.HandleResults()
	return err
}

type Player struct {
	colourBlind   bool
	hardMode      bool
	currStreak    int
	longestStreak int
	stats         [6]int
}

func setupDB() (*bolt.DB, error) {
	db, dbErr := bolt.Open("cliordle.db", 0600, nil)

	if dbErr != nil {
		return nil, fmt.Errorf("could not open db, %v", dbErr)
	}

	dbErr = db.Update(func(tx *bolt.Tx) error {
		root, bucketErr := tx.CreateBucketIfNotExists([]byte("DB"))
		if bucketErr != nil {
			return fmt.Errorf("could not create root bucket: %v", bucketErr)
		}
		_, bucketErr = root.CreateBucketIfNotExists([]byte("PLAYERDATA"))
		if bucketErr != nil {
			return fmt.Errorf("could not create player data bucket: %v", bucketErr)
		}
		return nil
	})
	if dbErr != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", dbErr)
	}
	fmt.Println("DB setup done")
	return db, nil
}

func initPlayer(db *bolt.DB) (*Player, error) {
	var player *Player
	err := db.View(func(tx *bolt.Tx) error {
		playerBytes := tx.Bucket([]byte("DB")).Get([]byte("PLAYERDATA"))
		var dbErr error
		dbErr = json.Unmarshal(playerBytes, player)
		return dbErr
	})
	return player, err
}

func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func play(db *bolt.DB, player *Player) error {
	answer, err := words.RandomWord()
	if err != nil {
		return err
	}
	currGame := Game{player, []string{}, answer, false}
	err = currGame.Play()
	return err
}

func settings(db *bolt.DB, player *Player) error {
	var err error
	return err
}

func stats(db *bolt.DB, player *Player) error {
	var err error
	return err
}

func manageCommands(db *bolt.DB, player *Player) error {
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
		err = play(db, player)
	} else if *action == "settings" {
		err = settings(db, player)
	} else {
		err = stats(db, player)
	}
	return err
}

func main() {
	db, dbErr := setupDB()

	if dbErr != nil {
		exitGracefully(dbErr)
	}

	defer db.Close()

	// display usage info when user enters --help option
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] \nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	player, err := initPlayer(db)

	// processing user command
	err = manageCommands(db, player)

	if err != nil {
		exitGracefully(err)
	}
}
