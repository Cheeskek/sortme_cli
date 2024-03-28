package main

import (
    "fmt"
    "os"
    "net/http"
    "net/url"
    "io"
    "errors"
    "strings"
    "strconv"
    "encoding/json"
    "bytes"
    "github.com/gorilla/websocket"
)

const (
    kApiUrl = "api.sort-me.org"
    kCacheDir = ".sm"
    kStatementsCache = "statements.json"
)

// json parsing types

type (
    Config struct {
        Token                 string     `json:"token"`
        Langs                 string     `json:"langs"`
    }

    Contest struct {
        Id                    int        `json:"id"`
        Name                  string     `json:"name"`
        Starts                uint64     `json:"starts"`
        Ends                  uint64     `json:"ends"`
        OrgName               string     `json:"org_name"`
        Running               bool       `json:"running"`
        RegistrationOpened    bool       `json:"registration_opened"`
        Ended                 bool       `json:"ended"`
    }

    // not all fields included, only nessesary
    Statements struct {
        Id                    uint       `json:"id"`
        Ends                  uint64     `json:"ends"`
        Langs struct {        
            Api               []string   `json:"api"`
            Extensions        [][]string `json:"extensions"`
            Verbose           []string   `json:"verbose"`
        }                                `json:"langs"`
        Name                  string     `json:"name"`
        ServerTime            uint64     `json:"server_time"`
        Status                string     `json:"status"`
        Tasks                 []Task     `json:"tasks"`
    }

    Task struct {
        Id                    uint       `json:"id"`
        Name                  string     `json:"name"`
        MainDescription       string     `json:"main_description"`
        InDescription         string     `json:"in_description"`
        OutDescription        string     `json:"out_description"`
        Category              uint64     `json:"category"`
        Difficulty            uint64     `json:"difficulty"`
        SolvedBy              uint64     `json:"solved_by"`
        Samples               []Sample   `json:"samples"`
        Comment               string     `json:"comment"`
        Admins                []uint     `json:"admins"`
        TestsUpdated          uint64     `json:"tests_updated"`
        TimeLimitMilliseconds uint64     `json:"time_limit_milliseconds"`
        MemoryLimitMegabytes  uint64     `json:"memory_limit_megabytes"`
        RatingSystemType      uint64     `json:"rating_system_type"`
    }

    Sample struct {
        In                    string     `json:"in"`
        Out                   string     `json:"out"`
    }

    Submission struct {
        Code                  string     `json:"code"`
        ContestId             uint       `json:"contest_id"`
        Lang                  string     `json:"lang"`
        TaskId                uint       `json:"task_id"`
    }

    Verdict struct {
        Compiled             bool        `json:"compiled"`
        CompilerLog          string      `json:"compiler_log"`
        ShownTest            uint        `json:"shown_test"`
        ShownVerdict         uint        `json:"shown_verdict"`
        ShownVerdictText     string      `json:"shown_verdict_text"`
        Subtasks []struct {
            FailedTests []struct {
                Milliseconds uint64     `json:"milliseconds"`
                N            uint64     `json:"n"`
                PartialScore uint64     `json:"partial_score"`
                Verdict      uint       `json:"verdict"`
                VerdictText  string     `json:"verdict_text"`
            }                           `json:"failed_tests"`
            Points           uint       `json:"points"`
            Skipped          bool       `json:"skipped"`
            WorstTime        uint64     `json:"worst_time"`
        }                               `json:"subtasks"`
        TotalPoints          uint       `json:"total_points"`
    }
)

// helper functions

func makeSortmeRequest(method string, req_url url.URL, body io.Reader, v any, config *Config) error {
    req, err := http.NewRequest(method, req_url.String(), body)
    if err != nil {
        return err
    }
    req.Header.Add("authorization", "Bearer " + config.Token)
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

    if res.StatusCode / 100 != 2 {
        return errors.New(fmt.Sprintf("Api error! Status code: %v\n", res.StatusCode))
    }

    body_bytes, err := io.ReadAll(res.Body)
    if err != nil {
        return err
    }
    err = json.Unmarshal(body_bytes, v)
    if err != nil {
        return err
    }

    return nil
}

func fetchAndParse(req_url url.URL, v any, config *Config) error {
    return makeSortmeRequest("GET", req_url, nil, v, config)
}

func sendSubmission(submission *Submission, v any, config *Config) error {
    body_bytes, err := json.Marshal(submission)
    if err != nil {
        return err
    }
    reader := bytes.NewReader(body_bytes)
    var req_url url.URL
    req_url.Scheme = "https"
    req_url.Host = kApiUrl
    req_url.Path = "/submit"
    return makeSortmeRequest("POST", req_url, reader, v, config)
}

func getConfig() (Config, error) {
    config_dir, err := os.UserConfigDir()
    if err != nil {
        return Config{}, err
    }

    config_bytes, err := os.ReadFile(config_dir + string(os.PathSeparator) + "sort-me" + string(os.PathSeparator) + "config.json")
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
    
    statements_file, err := os.ReadFile(kCacheDir + string(os.PathSeparator) + kStatementsCache)
    if os.IsNotExist(err) {
        return Statements{}, errors.New("Contest not chosen! Use \"contest\" command.")
    }
    if err != nil {
        return Statements{}, err
    }

    err = json.Unmarshal(statements_file, &statements)
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
    ext := filename_split[len(filename_split) - 1]

    for i := range statements.Langs.Api {
        for _, api_ext := range statements.Langs.Extensions[i] {
            if ext == api_ext {
                return statements.Langs.Api[i], nil
            }
        }
    }

    return "", errors.New("This langs extension is not supported")
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

func help() error {
    fmt.Printf("help message (to be added)\n")
    return nil
}

func contest(config *Config) error {
    if len(os.Args) >= 3 && os.Args[2] == "help" {
        fmt.Printf("contest help (to be added)")
        return nil
    }

    var req_url url.URL
    req_url.Scheme = "https"
    req_url.Host = kApiUrl
    req_url.Path = "/getUpcomingContests"
    var contests []Contest
    err := fetchAndParse(req_url, &contests, config)
    if err != nil {
        return err
    }

    fmt.Printf("Choose current contest:\n")
    for i := 0; i < len(contests); i++ {
        fmt.Printf("%d: %v | %v\n", i, contests[i].Name, contests[i].OrgName)
    }

    choice := ""
    var contest_ind int
    for choice == "" || err != nil || contest_ind >= len(contests)  {
        fmt.Scanln(&choice)
        contest_ind, err = strconv.Atoi(choice)
        if err != nil {
            fmt.Printf("Please input a number!\n")
            fmt.Printf("Choose current contest:\n")
        }
        if contest_ind >= len(contests) {
            fmt.Printf("Contest chosen is out of range!\n")
            fmt.Printf("Choose current contest:\n")
        }
    }

    
    contest_id := contests[contest_ind].Id
    
    req_url.Scheme = "https"
    req_url.Host = kApiUrl
    req_url.Path = "/getContestTasks"
    q := req_url.Query()
    q.Add("id", fmt.Sprint(contest_id))
    req_url.RawQuery = q.Encode()
    var statements Statements
    err = fetchAndParse(req_url, &statements, config)
    if err != nil {
        return err
    }

    statements.Id = uint(contest_id)

    statements_file, err := os.Create(kCacheDir + string(os.PathSeparator) + kStatementsCache)
    defer statements_file.Close()
    if err != nil {
        return err
    }
    json_bytes, err := json.Marshal(statements)
    if err != nil {
        return err
    }
    statements_file.Write(json_bytes)
    fmt.Printf("Current contest changed, statements written to %v\n", kCacheDir + string(os.PathSeparator) + kStatementsCache)


    return nil
}

func task() error {
    if len(os.Args) >= 3 && os.Args[2] == "help" {
        fmt.Printf("task help (to be added)")
        return nil
    } 

    statements, err := getStatements()
    if err != nil {
        return err
    }

    if (len(os.Args) >= 3 && os.Args[2] == "list") || len(os.Args) == 2 {
        for i, task := range statements.Tasks {
            fmt.Printf("%d: %v, Solved by: %v\n", i, task.Name, task.SolvedBy)
        }
        return nil
    } 

    task_ind, err := strconv.Atoi(os.Args[2])
    if err != nil {
        return err
    }
    if task_ind >= len(statements.Tasks) {
        return errors.New("Task index is out of range")
    }

    if len(os.Args) >= 4 && os.Args[3] == "sample" {
        if len(os.Args) == 4 {
            return errors.New("Sample index not given")
        }

        sample_ind, err := strconv.Atoi(os.Args[4])
        if err != nil {
            return err
        }
        if sample_ind >= len(statements.Tasks[task_ind].Samples) {
            return errors.New("Sample index out of range")
        }

        show_input := true
        show_output := true

        if len(os.Args) > 5 {
            show_input = false
            show_output = false
            for _, i := range os.Args[5] {
                switch i {
                case 'i':
                    show_input = true
                case 'o':
                    show_output = true
                }
            }
        }

        if show_input {
            fmt.Println(statements.Tasks[task_ind].Samples[sample_ind].In)
        }
        if show_output {
            fmt.Println(statements.Tasks[task_ind].Samples[sample_ind].Out)
        }

        return nil
    }

    show_legend := true
    show_in_desc := true
    show_out_desc := true
    show_comment := true

    if (len(os.Args) >= 5) {
        if os.Args[3] == "ignore" {
            for _, i := range os.Args[4] {
                switch i {
                case 'l':
                    show_legend = false
                case 'i':
                    show_in_desc = false
                case 'o':
                    show_out_desc = false
                case 'c':
                    show_comment = false
                }
            }
        } else if os.Args[3] == "only" {
            show_legend = false
            show_in_desc = false
            show_out_desc = false
            show_comment = false
            for _, i := range os.Args[4] {
                switch i {
                case 'l':
                    show_legend = true
                case 'i':
                    show_in_desc = true
                case 'o':
                    show_out_desc = true
                case 'c':
                    show_comment = true
                }
            }
        }
    }

    if show_legend {
        fmt.Printf("%v\n", statements.Tasks[task_ind].MainDescription)
    }
    if show_comment {
        fmt.Printf("%v\n\n", statements.Tasks[task_ind].Comment)
    } else {
        fmt.Printf("\n")
    }
    if show_in_desc {
        fmt.Printf("%v\n", statements.Tasks[task_ind].InDescription)
    }
    if show_out_desc {
        fmt.Printf("%v\n", statements.Tasks[task_ind].OutDescription)
    }

    return nil
}

func submit(config *Config) error {
    if (len(os.Args) >= 3 && os.Args[2] == "help") || len(os.Args) < 4 {
        fmt.Printf("submit help (to be added)")
        return nil
    } 

    statements, err := getStatements()
    if err != nil {
        return err
    }

    task_ind, err := strconv.Atoi(os.Args[2])
    if err != nil {
        return err
    }

    var code_bytes []byte

    var lang string

    if len(os.Args) == 5 {
        code_bytes, err = os.ReadFile(os.Args[3])
        if !doesLangExist(os.Args[5], &statements) {
            return errors.New("Passed lang is not supported by this contest")
        }
        lang = os.Args[4]
    } else if (isFilename(os.Args[3])) {
        code_bytes, err = os.ReadFile(os.Args[3])
        lang, err = getLang(os.Args[3], &statements)
        if err != nil {
            return err
        }
    } else {
        code_bytes, err = io.ReadAll(os.Stdin)
        if !doesLangExist(os.Args[3], &statements) {
            return errors.New("Passed lang is not supported by this contest")
        }
        lang = os.Args[3]
    }

    submission := Submission{
        string(code_bytes),
        statements.Id,
        lang,
        statements.Tasks[task_ind].Id,
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

    var req_url url.URL
    req_url.Scheme = "wss"
    req_url.Host = kApiUrl
    req_url.Path = "/ws/submission"
    q := req_url.Query()
    q.Add("id", fmt.Sprint(res.Id))
    q.Add("token", config.Token)
    req_url.RawQuery = q.Encode()
    c, _, err := websocket.DefaultDialer.Dial(
        req_url.String(),
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
            ch <- message;
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

// func rating(config *Config) error {
//     if len(os.Args) >= 3 && os.Args[2] == "help" {
//         fmt.Printf("rating help (to be added)")
//         return nil
//     } 
//
//     return nil
// }

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

    err = os.Mkdir(config_dir + string(os.PathSeparator) + "sort-me", os.ModePerm)
    if err != nil && !os.IsExist(err) {
        return Config{}, err
    }
    config_bytes, err := json.Marshal(config)
    if err != nil {
        return Config{}, err
    }
    err = os.WriteFile(config_dir + string(os.PathSeparator) + "sort-me" + string(os.PathSeparator) + "config.json", config_bytes, os.ModePerm)
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
    }
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    err = os.Mkdir(kCacheDir, os.ModePerm)
    if err != nil && !os.IsExist(err) {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if len(os.Args) <= 1 {
        err = help()
    } else {
        switch os.Args[1] {
        case "contest":
            err = contest(&config)
        case "task":
            err = task()
        case "submit":
            err = submit(&config)
        // case "rating":
        //     err = rating(&config)
        case "configure":
            if !already_configured {
                _, err = createConfig()
            }
        default:
            err = help()
        }
    }

    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
