package lib

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func ChangeContest(config *Config) error {
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

func PrintTask(taskInd int, only string, ignore string) error {
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

func PrintSample(taskInd int, sampleInd int, toShow string) error {
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

func Submit(taskInd int, filename string, lang string, config *Config) error {
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

func PrintRating(showByLabel bool, showTime bool, showAll bool, config *Config) error {
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
            if showTime {
                for _, res := range points.Results {
                    fmt.Printf("| %3d %2d:%02d:%02d ", res[0], res[1] / 3600, res[1] % 3600 / 60, res[1] % 60)
                }
            } else {
                for _, res := range points.Results {
                    fmt.Printf("| %3d ", res[0])
                }
            }
            fmt.Printf(" %d\n", points.Sum)
        }
        
        if !showAll {
            fmt.Printf("%d / %d pages\n", curPage, rating.Pages)
            fmt.Printf("%3d: you  ", rating.You.Place)
            if showTime {
                for _, res := range rating.You.Results {
                    fmt.Printf("| %3d %2d:%02d:%02d ", res[0], res[1] / 3600, res[1] % 3600 / 60, res[1] % 60)
                }
            } else {
                for _, res := range rating.You.Results {
                    fmt.Printf("| %3d ", res[0])
                }
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

func CreateConfig() (Config, error) {
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

func GetConfig() (Config, error) {
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

