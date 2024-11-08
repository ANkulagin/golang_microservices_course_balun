package main

import (
	"encoding/json"
	"fmt"
	desc "github.com/ANkulagin/golang_microservices_course_balun/week_1/grpc/pkg/note_v1"
	"github.com/brianvoe/gofakeit"
	"google.golang.org/protobuf/proto"
)

func main() {
	session := &desc.NoteInfo{
		Title:    gofakeit.BeerName(),
		Context:  gofakeit.BeerName(),
		Author:   gofakeit.Name(),
		IsPublic: gofakeit.Bool(),
	}

	dataJson, _ := json.Marshal(session)
	fmt.Printf("\n\ndataJson len %d byte \n%v\n", len(dataJson), dataJson)

	dataPb, _ := proto.Marshal(session)
	fmt.Printf("\n\ndataPb len %d byte \n%v\n", len(dataPb), dataPb)
}
