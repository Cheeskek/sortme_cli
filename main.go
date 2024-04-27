package main

import (
	"fmt"
	"github.com/Cheeskek/sortme_cli/internal/lib"
	"github.com/akamensky/argparse"
	"os"
)

type subcommand struct {
	command *argparse.Command
	handler func(*lib.Config) error
}

func getConfig() (*lib.Config, bool) {
	usedExisting := true

	config, err := lib.GetConfig()
	if os.IsNotExist(err) {
		usedExisting = false
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

	return &config, usedExisting
}

func addContestSubcommand(parser *argparse.Parser) subcommand {
	command := parser.NewCommand("c", "Choose contest from upcoming")

	handler := func(config *lib.Config) error {
		return lib.ChangeContest(config)
	}

	return subcommand{
		command,
		handler,
	}
}

func addTaskSubcommand(parser *argparse.Parser) subcommand {
	command := parser.NewCommand("t", "Display task description")

	ind := command.StringPositional(&argparse.Options{Default: "-1"})
	only := command.String("o", "only", nil)
	ignore := command.String("i", "ignore", nil)

	handler := func(config *lib.Config) error {
		taskNum, err := lib.TaskIndToInt(*ind)
		if err == nil {
			err = lib.PrintTask(taskNum, *only, *ignore)
		}
		return err
	}

	return subcommand{
		command,
		handler,
	}
}

func addSampleSubcommand(parser *argparse.Parser) subcommand {
	command := parser.NewCommand("s", "Display task sample")

	taskInd := command.StringPositional(&argparse.Options{Required: true})
	ind := command.Int("s", "sample", &argparse.Options{Default: 0})
	show := command.String("t", "type", &argparse.Options{Default: "io"})

	handler := func(config *lib.Config) error {
		taskNum, err := lib.TaskIndToInt(*taskInd)
		if err == nil {
			err = lib.PrintSample(taskNum, *ind, *show)
		}
		return err
	}

	return subcommand{
		command,
		handler,
	}
}

func addSubmitSubcommand(parser *argparse.Parser) subcommand {
	command := parser.NewCommand("S", "Submit solution")

	filename := command.StringPositional(nil)
	taskInd := command.StringPositional(&argparse.Options{Required: true})
	lang := command.String("l", "lang", nil)

	handler := func(config *lib.Config) error {
		taskNum, err := lib.TaskIndToInt(*taskInd)
		if err == nil {
			err = lib.Submit(taskNum, *filename, *lang, config)
		}
		return err
	}

	return subcommand{
		command,
		handler,
	}
}

func addConfigureSubcommand(parser *argparse.Parser, usedExistingConfig *bool) subcommand {
	command := parser.NewCommand("C", "Make config file")

	handler := func(config *lib.Config) error {
		if !*usedExistingConfig {
			_, err := lib.CreateConfig()
			return err
		}
		return nil
	}

	return subcommand{
		command,
		handler,
	}
}

func addRatingSubcommand(parser *argparse.Parser) subcommand {
	command := parser.NewCommand("r", "Get rating table")

	label := command.Flag("l", "label", &argparse.Options{Default: false})
	all := command.Flag("a", "all", &argparse.Options{Default: false})
	time := command.Flag("t", "time", &argparse.Options{Default: false})

	handler := func(config *lib.Config) error {
		return lib.PrintRating(*label, *time, *all, config)
	}

	return subcommand{
		command,
		handler,
	}
}

func main() {
	parser := argparse.NewParser("sortme_cli", "Surf Meat cli tool")

	config, usedExistingConfig := getConfig()

	subcommands := [...]subcommand{
		addContestSubcommand(parser),
		addTaskSubcommand(parser),
		addSampleSubcommand(parser),
		addSubmitSubcommand(parser),
		addConfigureSubcommand(parser, &usedExistingConfig),
		addRatingSubcommand(parser),
	}

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(parser.Usage(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}

	for _, cc := range subcommands {
		if cc.command.Happened() {
			if err := cc.handler(config); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	}
}
