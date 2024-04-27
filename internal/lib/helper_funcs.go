package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

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

func TaskIndToInt(taskInd string) (int, error) {
	taskNum, err := strconv.Atoi(taskInd)
	taskChar := []byte(taskInd)[0]
	if err != nil {
		if int('a') <= int(taskChar) && int(taskChar) <= int('z') {
			taskNum = int(taskChar) - int('a')
		} else if int('A') <= int(taskChar) && int(taskChar) <= int('Z') {
			taskNum = int(taskChar) - int('A')
		} else {
			return 0, fmt.Errorf("task format is not recognized")
		}
	}
	return taskNum, nil
}
