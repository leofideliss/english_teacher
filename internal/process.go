package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/leofideliss/english_teacher/pkg"
)

type Question struct {
    Text string `json:"text"`
}

type Answer struct {
    Model string `json:"model"`
    CreatedAt string `json:"created_at"`
    Response string `json:"response"`
    Done bool `json:"done"`
    DoneReason string `json:"done_reason"`
    Context []int `json:"context"`
    TotalDuration int64 `json:"total_duration"`
    LoadDuration int64 `json:"load_duration"`
    PromptEvalCount int `json:"prompt_eval_count"`
    PromptEvalDuration int64 `json:"prompt_eval_duration"`
    EvalCount int `json:"eval_count"`
    EvalDuration int64 `json:"eval_duration"`
}

type PartialAnswer struct {
	Response Message `json:"message"`
}

type Response struct {
    Status int `json:"status"`
    Success bool `json:"success"`
    Data string `json:"data"`
}

func consultLLM(request *http.Request) (io.ReadCloser , error) {
    payloadJson , errJson := makePayloadLLM(request.Body)
    if errJson != nil {
        return nil,errJson
    }

    return postLLM(payloadJson)
}

func bindResponseToanswer(resp io.ReadCloser, err error) Response{
    defer resp.Close()
    var answer Answer
    decoder := json.NewDecoder(resp)
    err = decoder.Decode(&answer)
    if err != nil {
        return Response{Status:http.StatusBadRequest,Data:err.Error(),Success:false}
    }
    return Response{Status:http.StatusOK,Data:answer.Response,Success:true}
}

func postLLM(payloadLLM []byte) (io.ReadCloser , error){
    resp, err := http.Post(os.Getenv("URL_LLM"), "application/json", bytes.NewBuffer(payloadLLM))
    if err != nil {
        return nil, err
    }
    return resp.Body , nil
}

func ExecuteQuestion(g *gin.Context){
    g.Header("Content-Type", "text/event-stream")
	g.Header("Cache-Control", "no-cache")
	g.Header("Connection", "keep-alive")
    var response string
    responseLLM , err := consultLLM(g.Request)
    if err != nil {
        g.SSEvent("error", fmt.Sprintf("Erro na requisição: %v", err))
    }
    
    reader := bufio.NewReader(responseLLM)

	g.Stream(func(w io.Writer) bool {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return false // Fim do stream
		}
        var partial PartialAnswer
        _ = json.Unmarshal([]byte(line), &partial)
        response += partial.Response.Content
		g.SSEvent("message", partial.Response.Content)
		return true // Continua lendo
	})
    handleHistory("assistant",response)
}

func saveQuestion(value string){
    redis.PushRedis(os.Getenv("CURRENT_CHAT"),value)
}

func handleHistory(agent , key string){  
    value := fmt.Sprintf("%s: %s", agent, key)
    saveQuestion(value)
}

func bindRequestToQuestion(body io.ReadCloser) (Question,error){
    var question Question
    decoder := json.NewDecoder(body)
    if err := decoder.Decode(&question); err != nil {
        return Question{}, err
    }
    handleHistory("user",question.Text)
    return question , nil
}

func makePayloadLLM(body io.ReadCloser)([]byte , error){
    question,err:= bindRequestToQuestion(body)
    result , _ := redis.RecuperarHistorico(os.Getenv("CURRENT_CHAT"),100)
    result = append(result,map[string]string{"role":"user","content":question.Text})

    stream , _ := strconv.ParseBool(os.Getenv("STREAM_LLM"))
    payload_llama := map[string]interface{}{
        "model": os.Getenv("MODEL_LLM"),
        "stream": stream,
        "messages": result,
    }

    jsonData, err := json.Marshal(payload_llama)
    if err != nil {
        return nil , err
    }

    return jsonData , nil
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func organizarHistorico(historico []map[string]string) ([]Message, error) {
    var mensagens []Message
    for _, entrada := range historico {
        if role, ok := entrada["role"]; ok {
            if content, ok := entrada["content"]; ok {
                mensagens = append(mensagens, Message{
                    Role:    role,
                    Content: content,
                })
            }
        }
    }
    return mensagens, nil
}
