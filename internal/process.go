package internal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Question struct {
    Text string `json:"text"`
}

type Answer struct {
    Model              string    `json:"model"`
    CreatedAt          string    `json:"created_at"`
    Response           string    `json:"response"`
    Done               bool      `json:"done"`
    DoneReason         string    `json:"done_reason"`
    Context            []int     `json:"context"`
    TotalDuration      int64     `json:"total_duration"`
    LoadDuration       int64     `json:"load_duration"`
    PromptEvalCount    int       `json:"prompt_eval_count"`
    PromptEvalDuration int64     `json:"prompt_eval_duration"`
    EvalCount          int       `json:"eval_count"`
    EvalDuration       int64     `json:"eval_duration"`
}

type Response struct {
    Status int `json:"status"`
    Success bool `json:"success"`
    Data string `json:"data"`
}

func consultLLM(request *http.Request) Response {
    payloadJson , errJson := prepareQuestion(request.Body)
    if errJson != nil {
        return Response{Status:http.StatusBadRequest,Data:errJson.Error(),Success:false}
    }
    
    resp, err := http.Post(os.Getenv("URL_LLM"), "application/json", bytes.NewBuffer(payloadJson))
	if err != nil {
        return Response{Status:http.StatusBadRequest,Data:errJson.Error(),Success:false}
	}
	defer resp.Body.Close()
    var answer Answer
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&answer)
	if err != nil {
        return Response{Status:http.StatusBadRequest,Data:errJson.Error(),Success:false}
    }
    return Response{Status:http.StatusOK,Data:answer.Response,Success:true}
}

func ExecuteQuestion(g *gin.Context){
    responseLLM  := consultLLM(g.Request)
    g.JSON(responseLLM.Status,responseLLM.Data)
}

func prepareQuestion(body io.ReadCloser) ([]byte,error){
	var question Question
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&question); err != nil {
		return nil, err
	}

    stream , _ := strconv.ParseBool(os.Getenv("STREAM_LLM"))
    payload_llama := map[string]interface{}{
		"model":  os.Getenv("MODEL_LLM"),
		"prompt":   question.Text,
		"stream": stream,
	}
    
    jsonData, err := json.Marshal(payload_llama)
	if err != nil {
		return nil , err
	}

    return jsonData , nil
}
