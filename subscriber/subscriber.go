package subscriber

import (
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "192.168.192.42:5050", "http service address")

type OracleRequestFormat struct {
	TYPE   string
	SOURCE string
	VALUE  string
}

type Response struct {
	REQUEST  string
	RESPOSNE map[string]string
}

func Subscribe(dataType string, source string, value string, dataChannel chan Response, quitChannel chan string) {
	//

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws/oracle"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	err = c.WriteJSON(&OracleRequestFormat{dataType, source, value})
	requestString := dataType + ":" + source + ":" + value

	for {
		//

		select {

		case <-time.After(5 * time.Second):
			log.Println("TIMEOUT")
			var r Response
			err = c.ReadJSON(&r)
			if err == nil {
				//

				dataChannel <- r

			} else {
				//

				return
			}

		case q := <-quitChannel:
			if q == requestString {
				//

				log.Println("TIME TO QUIT")
				return
			}
		}
	}
}
