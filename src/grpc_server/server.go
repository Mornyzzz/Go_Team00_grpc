package main

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	desc "goTeam00/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

func generateRandomValues() (float64, float64, string) {
	rand.Seed(time.Now().UnixNano())

	// Генерация случайного среднего значения в диапазоне [-10, 10]
	mean := rand.Float64()*20 - 10

	// Генерация случайного стандартного отклонения в диапазоне [0.3, 1.5]
	std := rand.Float64()*1.2 + 0.3

	uuid := uuid.New().String()

	return mean, std, uuid
}

func readValuesToFile(mean float64, std float64, uuid string) error {
	file, err := os.OpenFile("journal.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString("------------- NEW CONNECTION -------------\n")
	if err != nil {
		return err
	}

	logEntry := fmt.Sprintf("Mean: %f\t\t STD: %f\t\t UUID: %s\n", mean, std, uuid)
	_, err = file.WriteString(logEntry)
	if err != nil {
		return err
	}
	return nil
}

type server struct {
	desc.UnimplementedTransmiteServiceServer
}

func (s *server) StreamEntries(_ *empty.Empty, stream desc.TransmiteService_StreamEntriesServer) error {
	log.Println("StreamEntries called")
	mean, std, uuid := generateRandomValues()
	readValuesToFile(mean, std, uuid)

	for {
		entry := &desc.Entry{
			SessionId: uuid,
			Frequency: mean + rand.NormFloat64()*std,
			Timestamp: timestamppb.Now(),
		}
		err := stream.Send(entry)
		if err != nil {
			return err
		}
		log.Println(entry)

		time.Sleep(2 * time.Millisecond)
	}
}

func main() {
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterTransmiteServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
