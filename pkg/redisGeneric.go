package redis

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

type respositoryRedis struct{
    r *redis.Client
}

var ctx = context.Background()
var repository respositoryRedis

func init(){
    connectRedis()
}

func connectRedis() {
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",              
        DB:       0,                
    })
    repository.r = rdb
}

func PushRedis( key , value string) error{
    err := repository.r.RPush(ctx , key , value).Err()
    return err
}

func PopRedis( key string) ([]string , error){
    historico, err := repository.r.LRange(ctx, key, -10, -1).Result()
    if err != nil {
        return nil, err
    }

    return historico, nil
}

func RecuperarHistorico(key string, numMessages int64) ([]map[string]string, error) {
	// Recupera o histórico das últimas 'numMensagens' entradas
	historico, err := repository.r.LRange(ctx, key, 0, numMessages-1).Result()
	if err != nil {
		return nil, err
	}

	// Inicializa uma slice para armazenar o histórico no formato esperado
	var mensagens []map[string]string

	// Itera sobre as mensagens no histórico e formata o conteúdo
	for _, entrada := range historico {
		// Cada entrada tem o formato: "user: mensagem" ou "assistant: mensagem"
		parts := strings.SplitN(entrada, ": ", 2)
		if len(parts) == 2 {
			// Cria a estrutura de mensagem
			mensagem := map[string]string{
				"content": parts[1],    // Conteúdo da mensagem
				"role":    parts[0],    // Pode ser "user" ou "assistant"
			}
			mensagens = append(mensagens, mensagem)
		}
	}

	return mensagens, nil
}
