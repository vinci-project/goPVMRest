package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"flag"
	"goPVMRest/goVncRest"
	"log"
	"os"

	"github.com/mongodb/mongo-go-driver/mongo"
)

var mongoDB *mongo.Client

func connectToMongo() (err error) {
	//

	mongoDBHost, ok := os.LookupEnv("MONGO_HOST_TCP_ADDR")
	if !ok {
		//

		mongoDBHost = "192.168.192.42"
	}

	mongoDB, err = mongo.Connect(context.Background(), "mongodb://"+mongoDBHost+":27017", nil)
	if err != nil {
		//

		log.Fatal(err)
	}

	err = mongoDB.Ping(context.Background(), nil)
	if err != nil {
		//

		log.Fatalln("No connection to MONGO. ", err)
	}

	return
}

func main() {
	//

	privateKeyPtr := flag.String("privateKey",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"A private key associated with your node")

	ipPtr := flag.String("ip",
		"0.0.0.0",
		"Your external IP")

	flag.Parse()

	privateKey, err := hex.DecodeString(*privateKeyPtr)
	if err != nil {
		//

		panic("WRONG PRIVATE KEY")
	}

	if err := connectToMongo(); err != nil {
		//

		panic(err.Error())
	}

	defer mongoDB.Disconnect(context.Background())

	restServer.Start(mongoDB, *ipPtr, privateKey)

// 	consoleReader := bufio.NewReader(os.Stdin)
// 	log.Println("Enter text: ") // Just for better testing. Should be refactored for production
// 	consoleReader.ReadString('\n')
}
