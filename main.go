package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	API_URL          = "api.sort-me.org"
	CACHE_DIR        = ".sm"
	STATEMENTS_CACHE = "statements.json"
)

// json parsing types

type (
	Config struct {
		Token string `json:"token"`
		Langs string `json:"langs"`
	}

	Contest struct {
		Id                 int    `json:"id"`
		Name               string `json:"name"`
		Starts             uint64 `json:"starts"`
		Ends               uint64 `json:"ends"`
		OrgName            string `json:"org_name"`
		Running            bool   `json:"running"`
		RegistrationOpened bool   `json:"registration_opened"`
		Ended              bool   `json:"ended"`
	}

	// not all fields included, only nessesary
	Statements struct {
		Id    uint   `json:"id"`
		Ends  uint64 `json:"ends"`
		Langs struct {
			Api        []string   `json:"api"`
			Extensions [][]string `json:"extensions"`
			Verbose    []string   `json:"verbose"`
		} `json:"langs"`
		Name       string `json:"name"`
		ServerTime uint64 `json:"server_time"`
		Status     string `json:"status"`
		Tasks      []Task `json:"tasks"`
	}

	Task struct {
		Id                    uint     `json:"id"`
		Name                  string   `json:"name"`
		MainDescription       string   `json:"main_description"`
		InDescription         string   `json:"in_description"`
		OutDescription        string   `json:"out_description"`
		Category              uint64   `json:"category"`
		Difficulty            uint64   `json:"difficulty"`
		SolvedBy              uint64   `json:"solved_by"`
		Samples               []Sample `json:"samples"`
		Comment               string   `json:"comment"`
		Admins                []uint   `json:"admins"`
		TestsUpdated          uint64   `json:"tests_updated"`
		TimeLimitMilliseconds uint64   `json:"time_limit_milliseconds"`
		MemoryLimitMegabytes  uint64   `json:"memory_limit_megabytes"`
		RatingSystemType      uint64   `json:"rating_system_type"`
	}

	Sample struct {
		In  string `json:"in"`
		Out string `json:"out"`
	}

	Submission struct {
		Code      string `json:"code"`
		ContestId uint   `json:"contest_id"`
		Lang      string `json:"lang"`
		TaskId    uint   `json:"task_id"`
	}

	Verdict struct {
		Compiled             bool   `json:"compiled"`
		CompilerLog          string `json:"compiler_log"`
		ShownTest            uint   `json:"shown_test"`
		ShownVerdict         uint   `json:"shown_verdict"`
		ShownVerdictText     string `json:"shown_verdict_text"`
		Subtasks             []struct {
			FailedTests      []struct {
				Milliseconds uint64 `json:"milliseconds"`
				N            uint64 `json:"n"`
				PartialScore uint64 `json:"partial_score"`
				Verdict      uint   `json:"verdict"`
				VerdictText  string `json:"verdict_text"`
			}                       `json:"failed_tests"`
			Points           uint   `json:"points"`
			Skipped          bool   `json:"skipped"`
			WorstTime        uint64 `json:"worst_time"`
		}                           `json:"subtasks"`
		TotalPoints          uint   `json:"total_points"`
	}

    Rating struct {
        Frozen   bool     `json:"frozen"`
        Labels   []string `json:"labels"`
        Pages    uint     `json:"pages"`
        Table    []Points `json:"table"`
        You      Points   `json:"you"`
        YourPage uint     `json:"your_page"`
    }

    Points struct {
        Avatar  string   `json:"avatar"`
        Label   uint     `json:"label"`
        Login   string   `json:"login"`
        Place   uint64   `json:"place"`
        Results [][]int `json:"results"`
        Sum     uint64   `json:"sum"`
        Time    uint64   `json:"time"`
        Uid     uint     `json:"uid"`
    }
)

// helper functions

func makeSortmeRequest(method string, reqUrl url.URL, body io.Reader, v any, config *Config) error {
	req, err := http.NewRequest(method, reqUrl.String(), body)
	if err != nil {
		return err
	}
	req.Header.Add("authorization", "Bearer "+config.Token)
	req.Header.Add("accept-language", config.Langs)
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		message, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("api error: %v, message: %v", res.StatusCode, message)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodyBytes, v)
	if err != nil {
		return err
	}

	return nil
}

func fetchAndParse(reqUrl url.URL, v any, config *Config) error {
	return makeSortmeRequest("GET", reqUrl, nil, v, config)
}

func sendSubmission(submission *Submission, v any, config *Config) error {
	body_bytes, err := json.Marshal(submission)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(body_bytes)
	var reqUrl url.URL
	reqUrl.Scheme = "https"
	reqUrl.Host = API_URL
	reqUrl.Path = "/submit"
	return makeSortmeRequest("POST", reqUrl, reader, v, config)
}

func getConfig() (Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, err
	}

	config_bytes, err := os.ReadFile(configDir + string(os.PathSeparator) + "sort-me" + string(os.PathSeparator) + "config.json")
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(config_bytes, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func getStatements() (Statements, error) {
	var statements Statements

	statementsFile, err := os.ReadFile(CACHE_DIR + string(os.PathSeparator) + STATEMENTS_CACHE)
	if os.IsNotExist(err) {
		return Statements{}, fmt.Errorf("contest not chosen! Use \"contest\" command")
	}
	if err != nil {
		return Statements{}, err
	}

	err = json.Unmarshal(statementsFile, &statements)
	if err != nil {
		return Statements{}, err
	}

	return statements, nil
}

func isFilename(filename string) bool {
	for _, i := range filename {
		if i == '.' {
			return true
		}
	}
	return false
}

func getLang(filename string, statements *Statements) (string, error) {
	filename_split := strings.Split(filename, ".")
	ext := filename_split[len(filename_split)-1]

	for i := range statements.Langs.Api {
		for _, apiExt := range statements.Langs.Extensions[i] {
			if ext == apiExt {
				return statements.Langs.Api[i], nil
			}
		}
	}

	return "", fmt.Errorf("this langs extension is not supported")
}

func doesLangExist(lang string, statements *Statements) bool {
	for _, i := range statements.Langs.Api {
		if i == lang {
			return true
		}
	}
	return false
}

// functions for actions

func contest(config *Config) error {
	var reqUrl url.URL
	reqUrl.Scheme = "https"
	reqUrl.Host = API_URL
	reqUrl.Path = "/getUpcomingContests"
	var contests []Contest
	err := fetchAndParse(reqUrl, &contests, config)
	if err != nil {
		return err
	}

	fmt.Printf("Choose current contest:\n")
	for i := 0; i < len(contests); i++ {
		fmt.Printf("%d: %v | %v\n", i, contests[i].Name, contests[i].OrgName)
	}

	choice := ""
	var contestInd int
	for choice == "" || err != nil || contestInd >= len(contests) {
		fmt.Scanln(&choice)
		contestInd, err = strconv.Atoi(choice)
		if err != nil {
			fmt.Printf("Please input a number!\n")
			fmt.Printf("Choose current contest:\n")
		}
		if contestInd >= len(contests) {
			fmt.Printf("Contest chosen is out of range!\n")
			fmt.Printf("Choose current contest:\n")
		}
	}

	contest_id := contests[contestInd].Id

	reqUrl.Scheme = "https"
	reqUrl.Host = API_URL
	reqUrl.Path = "/getContestTasks"
	q := reqUrl.Query()
	q.Add("id", fmt.Sprint(contest_id))
	reqUrl.RawQuery = q.Encode()
	var statements Statements
	err = fetchAndParse(reqUrl, &statements, config)
	if err != nil {
		return err
	}

	statements.Id = uint(contest_id)

	statementsFile, err := os.Create(CACHE_DIR + string(os.PathSeparator) + STATEMENTS_CACHE)
	if err != nil {
		return err
	}
	defer statementsFile.Close()

	json_bytes, err := json.Marshal(statements)
	if err != nil {
		return err
	}
	statementsFile.Write(json_bytes)
	fmt.Printf("Current contest changed, statements written to %v\n", CACHE_DIR+string(os.PathSeparator)+STATEMENTS_CACHE)

	return nil
}

func task(taskInd int, only string, ignore string) error {
	statements, err := getStatements()
	if err != nil {
		return err
	}

	if taskInd == -1 {
		for i, task := range statements.Tasks {
			fmt.Printf("%d: %v, Solved by: %v\n", i, task.Name, task.SolvedBy)
		}
		return nil
	}

	if taskInd >= len(statements.Tasks) {
		return fmt.Errorf("Task index is out of range")
	}

	showLegend := true
	showInDesc := true
	showOutDesc := true
	showComment := true

	if ignore != "" {
		for _, i := range ignore {
			switch i {
			case 'l':
				showLegend = false
			case 'i':
				showInDesc = false
			case 'o':
				showOutDesc = false
			case 'c':
				showComment = false
			}
		}
	} else if only != "" {
		showLegend = false
		showInDesc = false
		showOutDesc = false
		showComment = false
		for _, i := range only {
			switch i {
			case 'l':
				showLegend = true
			case 'i':
				showInDesc = true
			case 'o':
				showOutDesc = true
			case 'c':
				showComment = true
			}
		}
	}

	if showLegend {
		fmt.Printf("%v\n", statements.Tasks[taskInd].MainDescription)
	}
	if showComment {
		fmt.Printf("%v\n\n", statements.Tasks[taskInd].Comment)
	} else {
		fmt.Printf("\n")
	}
	if showInDesc {
		fmt.Printf("%v\n", statements.Tasks[taskInd].InDescription)
	}
	if showOutDesc {
		fmt.Printf("%v\n", statements.Tasks[taskInd].OutDescription)
	}

	return nil
}

func sample(taskInd int, sampleInd int, toShow string) error {
	statements, err := getStatements()
	if err != nil {
		return err
	}

	if taskInd >= len(statements.Tasks) {
		return fmt.Errorf("Task index is out of range")
	}
	if sampleInd >= len(statements.Tasks[taskInd].Samples) {
		return fmt.Errorf("Sample index out of range")
	}

	showInput := false
	showOutput := false
	for _, i := range toShow {
		switch i {
		case 'i':
			showInput = true
		case 'o':
			showOutput = true
		}
	}

	if showInput {
		fmt.Println(statements.Tasks[taskInd].Samples[sampleInd].In)
	}
	if showOutput {
		fmt.Println(statements.Tasks[taskInd].Samples[sampleInd].Out)
	}

	return nil
}

func submit(taskInd int, filename string, lang string, config *Config) error {
	statements, err := getStatements()
	if err != nil {
		return err
	}

    if taskInd < 0 || taskInd > len(statements.Tasks) {
        return fmt.Errorf("task index out of range")
    }

	var codeBytes []byte

    if filename == "" {
		codeBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		if !doesLangExist(lang, &statements) {
			return fmt.Errorf("passed lang is not supported by this contest")
		}
	} else if lang == "" {
		codeBytes, err = os.ReadFile(filename)
		if err != nil {
			return err
		}
		lang, err = getLang(filename, &statements)
		if err != nil {
			return err
		}
    } else {
		codeBytes, err = os.ReadFile(filename)
		if err != nil {
			return err
		}
		if !doesLangExist(lang, &statements) {
			return fmt.Errorf("passed lang is not supported by this contest")
		}
    }

	submission := Submission{
		string(codeBytes),
		statements.Id,
		lang,
		statements.Tasks[taskInd].Id,
	}

	var res struct {
		Id uint `json:"id"`
	}

	err = sendSubmission(&submission, &res, config)
	if err != nil {
		return err
	}

	websock_header := http.Header{}
	websock_header.Add("accept-language", config.Langs)

	var reqUrl url.URL
	reqUrl.Scheme = "wss"
	reqUrl.Host = API_URL
	reqUrl.Path = "/ws/submission"
	q := reqUrl.Query()
	q.Add("id", fmt.Sprint(res.Id))
	q.Add("token", config.Token)
	reqUrl.RawQuery = q.Encode()
	c, _, err := websocket.DefaultDialer.Dial(
		reqUrl.String(),
		websock_header,
	)
	if err != nil {
		return err
	}
	defer c.Close()

	ch := make(chan []byte)

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error making connection with server: %v\n", err)
				return
			}
			ch <- message
			if message[0] == byte('{') {
				return
			}
		}
	}()

	for {
		message := <-ch
		if message[0] == '{' {
			var verdict Verdict
			json.Unmarshal(message, &verdict)
			fmt.Printf("%d %v\n", verdict.TotalPoints, verdict.ShownVerdictText)
			break
		} else {
			fmt.Printf("%v\n", string(message))
		}
	}

	return nil
}

func rating(showByLabel bool, showAll bool, config *Config) error {
	statements, err := getStatements()
	if err != nil {
		return err
	}

    curPage := uint(1)
    curLabel := 0
    var rating Rating

    var reqUrl url.URL
    reqUrl.Scheme = "https"
    reqUrl.Host = API_URL
    reqUrl.Path = "getContestTable"
	q := reqUrl.Query()
    q.Add("contestid", fmt.Sprint(statements.Id))
	q.Add("page", "1")
	q.Add("label", "0")
	reqUrl.RawQuery = q.Encode()
    makeSortmeRequest("GET", reqUrl, nil, &rating, config)

    if showByLabel {
        fmt.Printf("Choose label:\n")
        for ind, label := range rating.Labels {
            fmt.Printf("%d: %s\n", ind + 1, label)
        }
        var choice string
        fmt.Scanln(&choice)
        curLabel, err = strconv.Atoi(choice)
        if err != nil {
            return err
        }
        q := reqUrl.Query()
        q.Set("label", fmt.Sprint(curLabel))
        reqUrl.RawQuery = q.Encode()
        makeSortmeRequest("GET", reqUrl, nil, &rating, config)
    }

    for {
        for _, points := range rating.Table {
            fmt.Printf("%3d: %30s ", points.Place, points.Login)
            for _, res := range points.Results {
                fmt.Printf("| %3d %2d:%02d:%02d ", res[0], res[1] / 3600, res[1] % 3600 / 60, res[1] % 60)
            }
            fmt.Printf(" %d\n", points.Sum)
        }
        
        if !showAll {
            fmt.Printf("%d / %d pages\n", curPage, rating.Pages)
            fmt.Printf("%3d: you  ", rating.You.Place)
            for _, res := range rating.You.Results {
                fmt.Printf("| %3d %2d:%02d:%02d ", res[0], res[1] / 3600, res[1] % 3600 / 60, res[1] % 60)
            }
            fmt.Printf(" %d\n", rating.You.Sum)
        }

        quit := false
        if showAll {
            curPage++
            if curPage == rating.Pages + 1 {
                quit = true
            }
        } else {
            validChoice := false
            var choice string
            for !validChoice {
                fmt.Scanln(&choice)
                validChoice = true
                switch choice {
                case "+":
                    curPage = curPage % rating.Pages + 1
                case "-":
                    curPage = (curPage + rating.Pages - 2) % rating.Pages + 1
                case "q":
                    quit = true
                default:
                    fmt.Println("Wrong option: (+/-/q)")
                    validChoice = false
                }
            }
        }
        if quit {
            break
        }
        q := reqUrl.Query()
        q.Set("page", fmt.Sprint(curPage))
        reqUrl.RawQuery = q.Encode()
        makeSortmeRequest("GET", reqUrl, nil, &rating, config)
    }

    return nil
}

func createConfig() (Config, error) {
	config_dir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, err
	}

	var config Config
	fmt.Printf("Please paste your API key\n")
	fmt.Scanln(&config.Token)
	fmt.Printf("Please put your preffered languages (Example: \"ru, en-US\" without quotes)\n")
	fmt.Scanln(&config.Langs)
	fmt.Printf("Now creating your config file, do not hang up!\n")

	err = os.Mkdir(config_dir+string(os.PathSeparator)+"sort-me", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return Config{}, err
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		return Config{}, err
	}
	err = os.WriteFile(config_dir+string(os.PathSeparator)+"sort-me"+string(os.PathSeparator)+"config.json", configBytes, os.ModePerm)
	if err != nil {
		return Config{}, err
	}

	fmt.Printf("Config file created at " + config_dir + string(os.PathSeparator) + "sort-me" + string(os.PathSeparator) + "config.json" + ". Happy SortMeing!\n")

	return config, nil
}

func main() {
	var err error = nil

	already_configured := false

	config, err := getConfig()
	if os.IsNotExist(err) {
		fmt.Printf("Config not found... Let's create it!\n")
		config, err = createConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = os.Mkdir(CACHE_DIR, os.ModePerm)
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

	err = parser.Parse(os.Args)
	if err != nil {
        fmt.Println(parser.Usage(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}

	if contestCom.Happened() {
		err = contest(&config)
	} else if taskCom.Happened() {
		err = task(*taskInd, *taskOnly, *taskIgnore)
	} else if sampleCom.Happened() {
		err = sample(*sampleTaskInd, *sampleInd, *sampleShow)
	} else if submitCom.Happened() {
		err = submit(*submitTaskInd, *submitFilename, *submitLang, &config)
	} else if configureCom.Happened() {
		if !already_configured {
			_, err = createConfig()
		}
	} else if ratingCom.Happened() {
        rating(*ratingLabel, *ratingAll, &config)
    }

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
