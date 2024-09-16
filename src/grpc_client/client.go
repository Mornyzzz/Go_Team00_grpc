package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"goTeam00/postgres"
	desc "goTeam00/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

func readValuesToFile(mean float64, std float64, count int) error {
	file, err := os.OpenFile("../grpc_server/journal.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	logEntry := fmt.Sprintf("Processed: %d\t\t\t\t\t PredictMean: %f\t\t PredictSTD: %f\n", count, mean, std)
	_, err = file.WriteString(logEntry)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	kFlag := flag.Int("k", 0, "for anomaly")
	flag.Parse()
	if *kFlag == 0 {
		log.Fatal("kFlag is zero or not set")
	}
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := desc.NewTransmiteServiceClient(conn)
	stream, err := c.StreamEntries(context.Background(), &empty.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var sum, mean, std float64
	var count int

	// подсчет
	pool := sync.Pool{}
	fmt.Println("Mean/STD reconstruction...")
	for i := 0; i < 180; i++ {
		responce, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		sum += responce.Frequency
		count++

		pool.Put(responce.Frequency)
		mean = sum / float64(count)

		var squaredDiff float64
		newPool := sync.Pool{}
		for value := pool.Get(); value != nil; value = pool.Get() {
			diff := value.(float64) - mean
			squaredDiff += diff * diff

			newPool.Put(value)
		}
		for value := newPool.Get(); value != nil; value = newPool.Get() {
			pool.Put(value)
		}
		std = math.Sqrt(squaredDiff / float64(count))

		readValuesToFile(mean, std, count)
	}
	//аномалии
	fmt.Println("Detecting anomalies...")
	var anomaly_count int
	dbase := postgres.GetDB()
	for {
		count++
		responce, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		maxDeviation := math.Abs(float64(*kFlag) * std)
		if responce.Frequency < mean-maxDeviation || responce.Frequency > mean+maxDeviation {
			anomaly_count++

			result := dbase.Create(postgres.Record{
				responce.SessionId,
				responce.Frequency,
				time.Unix(responce.Timestamp.GetSeconds(), int64(responce.Timestamp.GetNanos())).UTC(),
			})

			if result.Error != nil {
				log.Fatal(result.Error)
			}
			fmt.Printf("anomaly (%d/%d)\n", anomaly_count, count)
		}
	}

}
