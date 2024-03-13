package rpc

import (
	"log"

	"github.com/Jourloy/go-metrics-collector/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Connect() proto.MetricServiceClient {
	// устанавливаем соединение с сервером
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// получаем переменную интерфейсного типа UsersClient,
	// через которую будем отправлять сообщения
	c := proto.NewMetricServiceClient(conn)
	return c
}
