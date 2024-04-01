package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
    "github.com/Cheeskek/sortme_cli/internal/lib"
)

func main() {
	var err error = nil

	already_configured := false

	config, err := lib.GetConfig()
	if os.IsNotExist(err) {
		fmt.Printf("Config not found... Let's create it!\n")
		config, err = lib.CreateConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = os.Mkdir(lib.CACHE_DIR, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	err = nil

	parser := argparse.NewParser("sortme_cli", "Spores moomins cli tool")

	contestCom := parser.NewCommand("contest", "Choose contest from upcoming")

	taskCom := parser.NewCommand("task", "Display task description")
	taskInd := taskCom.IntPositional(&argparse.Options{Default: -1})
    taskOnly := taskCom.String("o", "only", nil)
    taskIgnore := taskCom.String("i", "ignore", nil)
    
	sampleCom := parser.NewCommand("sample", "Display task sample")
    sampleTaskInd := sampleCom.IntPositional(&argparse.Options{Required: true})
    sampleInd := sampleCom.Int("s", "sample", &argparse.Options{Default: 0})
    sampleShow := sampleCom.String("t", "type", &argparse.Options{Default: "io"})

	submitCom := parser.NewCommand("submit", "Submit solution")
    submitTaskInd := submitCom.IntPositional(&argparse.Options{Required: true})
    submitFilename := submitCom.StringPositional(nil)
    submitLang := submitCom.String("l", "lang", nil)

	configureCom := parser.NewCommand("configure", "Make conifg file")

    ratingCom := parser.NewCommand("rating", "Get rating table")
    ratingLabel := ratingCom.Flag("l", "label", &argparse.Options{Default: false})
    ratingAll := ratingCom.Flag("a", "all", &argparse.Options{Default: false})
    ratingTime := ratingCom.Flag("t", "time", &argparse.Options{Default: false})

	err = parser.Parse(os.Args)
	if err != nil {
        fmt.Println(parser.Usage(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}

	if contestCom.Happened() {
		err = lib.ChangeContest(&config)
	} else if taskCom.Happened() {
		err = lib.PrintTask(*taskInd, *taskOnly, *taskIgnore)
	} else if sampleCom.Happened() {
		err = lib.PrintSample(*sampleTaskInd, *sampleInd, *sampleShow)
	} else if submitCom.Happened() {
		err = lib.Submit(*submitTaskInd, *submitFilename, *submitLang, &config)
	} else if ratingCom.Happened() {
        lib.PrintRating(*ratingLabel, *ratingTime, *ratingAll, &config)
	} else if configureCom.Happened() {
		if !already_configured {
			_, err = lib.CreateConfig()
		}
    }

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
