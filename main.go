package main

import (
    "fmt"
    "os"
    "net/http"
    "io"
    "errors"
    "strconv"
    "encoding/json"
)

const (
    kApiUrl = "https://api.sort-me.org"
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
)

// helper functions

func fetchAndParse(payload string, v any, config *Config) error {
    req, err := http.NewRequest("GET", kApiUrl + payload, nil)
    if err != nil {
        return err
    }
    req.Header.Add("authorization", "Bearer " + config.Token)
    req.Header.Add("accept-language", config.Langs)

    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
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

func getConfig() (Config, error) {
    config_dir, err := os.UserConfigDir()
    if err != nil {
        return Config{}, err
    }
    
    config_bytes, err := os.ReadFile(config_dir + "/sort-me/config.json")
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

    var contests []Contest
    err := fetchAndParse("/getUpcomingContests", &contests, config)
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
    var statements Statements
    err = fetchAndParse("/getContestTasks?id=" + fmt.Sprint(contest_id), &statements, config)
    if err != nil {
        return err
    }

    statements_file, err := os.Create(kCacheDir + "/" + kStatementsCache)
    defer statements_file.Close()
    if err != nil {
        return err
    }
    json_bytes, err := json.Marshal(statements)
    if err != nil {
        return err
    }
    statements_file.Write(json_bytes)
    fmt.Printf("Current contest changed, statements written to %v\n", kCacheDir + "/" + kStatementsCache)


    return nil
}

func task(config *Config) error {
    if len(os.Args) >= 3 && os.Args[2] == "help" {
        fmt.Printf("task help (to be added)")
        return nil
    } 

    show_legend := true
    show_in_desc := true
    show_out_desc := true
    show_comment := true

    if (len(os.Args) >= 4) {
        if os.Args[2] == "ignore" {
            for _, i := range os.Args[3] {
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
        } else if os.Args[2] == "only" {
            show_legend = false
            show_in_desc = false
            show_out_desc = false
            show_comment = false
            for _, i := range os.Args[3] {
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

    var statements Statements
    statements_file, err := os.ReadFile(kCacheDir + "/" + kStatementsCache)
    if os.IsNotExist(err) {
        fmt.Printf("Contest not chosen!")
        err = contest(config)
    }
    if err != nil {
        return err
    }
    json.Unmarshal(statements_file, &statements)

    if len(os.Args) < 2 {
        return errors.New("Task number not given!")
    }


    if (len(os.Args) >= 3 && os.Args[2] == "list") || len(os.Args) == 2 {
        for i, task := range statements.Tasks {
            fmt.Printf("%d: %v, Solved by: %v\n", i, task.Name, task.SolvedBy)
        }
        return nil
    } 

    task_ind, err := strconv.Atoi(os.Args[len(os.Args) - 1])
    if err != nil {
        return err
    }
    if task_ind >= len(statements.Tasks) {
        return errors.New("Tasks number is out of range")
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

// func submit(config *Config) error {
//     if len(os.Args) >= 3 && os.Args[2] == "help" {
//         fmt.Printf("submit help (to be added)")
//         return nil
//     } 
//
//     return nil
// }

// func rating(config *Config) error {
//     if len(os.Args) >= 3 && os.Args[2] == "help" {
//         fmt.Printf("rating help (to be added)")
//         return nil
//     } 
//
//     return nil
// }

func configure() (Config, error) {
    config_dir, err := os.UserConfigDir()
    if err != nil {
        return Config{}, err
    }

    var config Config
    fmt.Printf("Please paste your API key\n")
    fmt.Scanln(&config.Token)
    fmt.Println(config.Token)
    fmt.Printf("Please put your preffered languages (Example: \"ru, en-US\" without quotes)\n")
    fmt.Scanln(&config.Langs)
    fmt.Println(config.Langs)
    fmt.Printf("Now creating your config file, do not hang up!\n")

    err = os.Mkdir(config_dir + "/sort-me", os.ModePerm)
    if err != nil && !os.IsExist(err) {
        return Config{}, err
    }
    config_bytes, err := json.Marshal(config)
    if err != nil {
        return Config{}, err
    }
    err = os.WriteFile(config_dir + "/sort-me/config.json", config_bytes, os.ModePerm)
    if err != nil {
        return Config{}, err
    }

    fmt.Printf("Config file created at " + config_dir + "/sort-me/config.json" + ". Happy SortMeing!\n")

    return config, nil
}

func main() {
    var err error = nil

    already_configured := false

    config, err := getConfig()
    if os.IsNotExist(err) {
        fmt.Printf("Config not found... Let's create it!\n")
        config, err = configure()
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
            err = task(&config)
        // case "submit":
        //     err = submit(&config)
        // case "rating":
        //     err = rating(&config)
        case "configure":
            if !already_configured {
                _, err = configure()
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
