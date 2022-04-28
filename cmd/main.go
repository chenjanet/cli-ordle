package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
	"github.com/j985chen/cli-ordle/words"
)

type Game struct {
	player       *Player
	numGuesses   int
	wordsGuessed []string
	answer       string
	board        string
}

func (g *Game) ProcessGuess(guess string) error {
	var err error
	return err
}

func (g *Game) PrintBoard() {

}

type Stats struct {
	numOnes   int
	numTwos   int
	numThrees int
	numFours  int
	numFives  int
	numSixes  int
}

type Player struct {
	colourBlind   bool
	hardMode      bool
	currStreak    int
	longestStreak int
	stats         Stats
}

func setupDB() (*bolt.DB, error) {
	db, dbErr := bolt.Open("gotodo.db", 0600, nil)

	if dbErr != nil {
		return nil, fmt.Errorf("could not open db, %v", dbErr)
	}

	dbErr = db.Update(func(tx *bolt.Tx) error {
		root, bucketErr := tx.CreateBucketIfNotExists([]byte("DB"))
		if bucketErr != nil {
			return fmt.Errorf("could not create root bucket: %v", bucketErr)
		}
		_, bucketErr = root.CreateBucketIfNotExists([]byte("TODOENTRIES"))
		if bucketErr != nil {
			return fmt.Errorf("could not create todo entry bucket: %v", bucketErr)
		}
		return nil
	})
	if dbErr != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", dbErr)
	}
	return db, nil
}

func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func play(db *bolt.DB, player *Player) error {
	var err error
	reader := bufio.NewReader(os.Stdin)
	answer, err := words.RandomWord()
	currGame := Game{player, 0, []string{}, answer, ""}
	for i := 0; i < 6; i++ {
		fmt.Println("Enter your first guess:")
		var guess string
		guess, err = reader.ReadString('\n')
		err = currGame.ProcessGuess(guess)
		currGame.PrintBoard()
	}
	return err
}

func settings(db *bolt.DB) error {
	var err error
	return err
}

func stats(db *bolt.DB) error {
	var err error
	return err
}

func initPlayer(db *bolt.DB) (*Player, error) {
	return &Player{}, nil
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
		err = settings(db)
	} else {
		err = stats(db)
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
		fmt.Printf("Usage: %s [options] <item to add or complete\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	player, err := initPlayer(db)

	// processing user command
	err = manageCommands(db, player)

	if err != nil {
		exitGracefully(err)
	}
}
