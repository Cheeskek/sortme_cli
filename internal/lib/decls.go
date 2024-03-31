package lib

const (
	API_URL          = "api.sort-me.org"
	CACHE_DIR        = ".sm"
	STATEMENTS_CACHE = "statements.json"
)

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
		Id             uint       `json:"id"`
		Ends           uint64     `json:"ends"`
		Langs struct {
			Api        []string   `json:"api"`
			Extensions [][]string `json:"extensions"`
			Verbose    []string   `json:"verbose"`
		}                         `json:"langs"`
		Name           string     `json:"name"`
		ServerTime     uint64     `json:"server_time"`
		Status         string     `json:"status"`
		Tasks          []Task     `json:"tasks"`
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
		Samples               []struct {
            In                string   `json:"in"`
            Out               string   `json:"out"`
        }                              `json:"samples"`
		Comment               string   `json:"comment"`
		Admins                []uint   `json:"admins"`
		TestsUpdated          uint64   `json:"tests_updated"`
		TimeLimitMilliseconds uint64   `json:"time_limit_milliseconds"`
		MemoryLimitMegabytes  uint64   `json:"memory_limit_megabytes"`
		RatingSystemType      uint64   `json:"rating_system_type"`
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

