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
    api_url = "https://api.sort-me.org"
    cache_dir = ".sm"
    statements_cache = "statements.json"
    auth = "Bearer <token>"
    languages = "ru,en-US"
)

// json parsing types

type (
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

func fetchAndParse(payload string, v any) error {
    req, err := http.NewRequest("GET", api_url + payload, nil)
    if err != nil {
        return err
    }
    req.Header.Add("authorization", auth)
    req.Header.Add("accept-language", languages)

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

// functions for actions

func help() error {
    fmt.Printf("help message (to be added)\n")
    return nil
}

func contest() error {
    if len(os.Args) >= 3 && os.Args[2] == "help" {
        fmt.Printf("contest help (to be added)")
        return nil
    }

    var contests []Contest
    err := fetchAndParse("/getUpcomingContests", &contests)
    if err != nil {
        return err
    }

    fmt.Printf("Choose current contest:\n")
    for i := 0; i < len(contests); i++ {
        fmt.Printf("%d: %v | %v\n", i, contests[i].Name, contests[i].OrgName)
    }

    var choice string
    fmt.Scanln(&choice)

    contest_ind, err := strconv.Atoi(choice)
    if err != nil {
        return err
    }

    contest_id := contests[contest_ind].Id
    var statements Statements
    err = fetchAndParse("/getContestTasks?id=" + fmt.Sprint(contest_id), &statements)
    if err != nil {
        return err
    }

    statements_file, err := os.Create(cache_dir + "/" + statements_cache)
    defer statements_file.Close()
    if err != nil {
        return err
    }
    json_bytes, err := json.Marshal(statements)
    if err != nil {
        return err
    }
    statements_file.Write(json_bytes)
    fmt.Printf("Current contest changed, statements written to %v\n", cache_dir + "/" + statements_cache)

    return nil
}

func task() error {
    if len(os.Args) >= 3 && os.Args[2] == "help" {
        fmt.Printf("task help (to be added)")
        return nil
    } 

    return nil
}

func submit() error {
    if len(os.Args) >= 3 && os.Args[2] == "help" {
        fmt.Printf("submit help (to be added)")
        return nil
    } 

    return nil
}

func rating() error {
    if len(os.Args) >= 3 && os.Args[2] == "help" {
        fmt.Printf("rating help (to be added)")
        return nil
    } 

    return nil
}

func main() {
    var err error = nil

    err = os.Mkdir(cache_dir, os.ModePerm)
    if err != nil && !os.IsExist(err) {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if len(os.Args) <= 1 {
        err = help()
    } else {
        switch os.Args[1] {
        case "contest":
            err = contest()
        case "task":
            err = task()
        case "submit":
            err = submit()
        case "rating":
            err = rating()
        default:
            err = help()
        }
    }

    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
