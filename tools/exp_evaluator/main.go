package main

import (
	"os"
	"flag"
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/influenzanet/study-service/pkg/types"
)

type ImportStruct struct {
	State types.ParticipantState`json:"state"`
	Rules []types.Expression `json:"rules"`
	Data []types.SurveyResponse `json:"responses"`	
}

type EvalResult struct {
	Index int `json:"index"`
	Data studyengine.ActionData`json:"result"`
	Error string `json:"error"`
}

func Evaluate(input ImportStruct) []EvalResult {
	instanceID := "dummy"
	studyKey := "dummy"
	
	dbService := MemoryDBService{Data: input.Data}
	
	event := types.StudyEvent{
		InstanceID:                            instanceID,
		StudyKey:                              studyKey,
		ParticipantIDForConfidentialResponses: "",
	}

	actionData := studyengine.ActionData{
		PState:          input.State,
		ReportsToCreate: map[string]types.Report{},
	}

	results := make([]EvalResult, 0)

	for index, rule := range input.Rules {
		newState, err := studyengine.ActionEval(rule, actionData, event, studyengine.ActionConfigs{
			DBService:            dbService ,
			ExternalServiceConfigs: nil,
		})

		r := EvalResult{
			Index: index,
			Data: newState,
		}
		if(err != nil) {
			r.Error =  fmt.Sprintf("%s", err)
		}

		results = append(results, r )

	}
	return results
}

func readImportStructFromJSON(filename string) ImportStruct {
	content, err := os.ReadFile(filename)
	if err != nil {
		logger.Error.Fatalf("Failed to read test-file: %s - %v", filename, err)
	}
	var input ImportStruct
	err = json.Unmarshal(content, &input)
	if err != nil {
		logger.Error.Fatal(err)
	}
	return input
}

func saveResultJSON(results []EvalResult, filename string) {
	file, _ := json.Marshal(results)
	err := os.WriteFile(filename, file, 0644)
	if err != nil {
		logger.Error.Fatal(err)
	}
}

func startServer(port int) {
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if(r.Method == "POST") {
			handlePost(w, r)
		} else {
			http.Error(w, "Only post please", http.StatusBadRequest)
		}
	})
	err := srv.ListenAndServe()
	fmt.Println(err)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	var input ImportStruct
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
	 http.Error(w, err.Error(), http.StatusBadRequest)
	 return
	}
	results := Evaluate(input)
	json.NewEncoder(w).Encode(results)
}

type Options struct {
	Input string
	Output string
	server int
}


func handleFlags() Options {
	inputJSON := flag.String("input", "", "path and name of the input file that should be converted")
	outputJSON := flag.String("output", "", "path and name of the input file that should be converted")
	server := flag.Int("server", 0, "Use server port (0 = disable)")
	flag.Parse()
	
	inputFile := *inputJSON
	outputFile := *outputJSON
	serverPort :=  *server
	
	if(inputFile != "") {
		if(serverPort > 0) {
			logger.Error.Fatal("Cannot use -server with -input")
		}
		if(outputFile == "") {
			logger.Error.Fatal("-output required with -input")
		}
	} 
	
	return Options{
		Input: inputFile,
		Output: outputFile,
		server:serverPort,
	}
}



func main() {

	opts := handleFlags()

	fmt.Println(opts)

	if(opts.Input != "") {
		fmt.Println("Using files")
		
		input := readImportStructFromJSON(opts.Input)
		r := Evaluate(input)
		saveResultJSON(r, opts.Output)
	}

	if(opts.server > 0) {
		fmt.Println("Starting server")
		startServer(opts.server)
	}

}