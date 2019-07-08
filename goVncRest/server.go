package restServer

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"goPVMRest/goVncRest/tools"
	"goPVMRest/helpers"
	"goPVMRest/subscriber"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	secp "github.com/vncsphere-foundation/secp256k1-go"
	"github.com/valyala/fasthttp"
)

var mongoDB *mongo.Client
var privateKey []byte
var publicKey []byte
var pvmHost string

type entryPoint struct {
	NAME string
}

type arg struct {
	TYPE  string
	VALUE string
}

type op struct {
	TYPE string
}

type body struct {
	NAME string
}

type comparator struct {
	TYPE  string
	VALUE string
}

type startRule struct {
	LIBRARY     string
	FUNCTION    string
	ARGUMENTS   []arg
	OPS         []op
	COMPARATORS []comparator
	BODY        []body
}

type anal struct {
	ENTRY_POINTS []entryPoint
	START_RULES  []startRule
}

type dapp struct {
	TT        string
	TST       string
	CODE      string
	ANALYSIS  anal
	SENDER    string
	PRICE     string
	SIGNATURE string
}

type dappCall struct {
	TT        string
	TST       string
	ID        string
	SENDER    string
	FUNC      []string
	ARGS      []string
	SIGNATURE string
}

type dappCallForSign struct {
	TT     string
	TST    string
	ID     string
	SENDER string
	FUNC   []string
	ARGS   []string
}

type callCondition struct {
	CONDITION string
	TYPE      string
	VALUE     string
	FUNC      []string
	ID        string
	REQUEST   string
	INWORK    bool
}

type startstop struct {
	SIGNATURE string
	SENDER    string
	ACTION    string
	SMS       string
	TST       string
}

var allCallConditions map[string]callCondition
var allDataSubscribes map[string]bool
var dataChannel chan subscriber.Response
var quitChannel chan string

func fastHTTPRawHandler(ctx *fasthttp.RequestCtx) {
	if string(ctx.Method()) == "GET" {
		//

		switch string(ctx.Path()) {

		case "/status":
			tools.MakeResponse(helpers.StatusOk, ctx)
			return
		}
	}

	if string(ctx.Method()) == "POST" {
		//

		switch string(ctx.Path()) {

		case "/scheduler/newBlock":
			//
			log.Println("ENDPOINT CALL")
			args := ctx.QueryArgs()
			for errNum, v := range helpers.PostNewSCFields {
				//

				if !args.Has(v) {
					//

					tools.MakeResponse(errNum, ctx)
					return
				}
			}

			bHeight, _ := strconv.ParseInt(string(args.Peek("BHEIGHT")), 10, 64)
			log.Println(bHeight)
			db := mongoDB.Database("pvmStemNode")
			collection := db.Collection("bchain")
			filter := bson.D{{"BHEIGHT", bHeight}}
			result := bson.Raw{}
			result, err := collection.FindOne(context.Background(), filter).DecodeBytes()
			if err != nil {
				//

				tools.MakeResponse(helpers.StatusDataNotFound, ctx)
				return
			}

			var dapps []dapp
			val := result.Lookup("DAPP").String()
			json.Unmarshal([]byte(val), &dapps)

			for _, uniqueDapp := range dapps {
				//

				log.Println(uniqueDapp.ANALYSIS)
				startRules := uniqueDapp.ANALYSIS.START_RULES

				for _, startRule := range startRules {
					//

					if startRule.LIBRARY == "smartExchange" {
						//

						var condition callCondition
						condition.INWORK = true
						condition.CONDITION = startRule.OPS[0].TYPE
						condition.TYPE = startRule.COMPARATORS[0].TYPE
						condition.VALUE = startRule.COMPARATORS[0].VALUE
						for i := range startRule.BODY {
							//

							condition.FUNC = append(condition.FUNC, startRule.BODY[i].NAME)
						}

						condition.ID = uniqueDapp.SIGNATURE

						if startRule.FUNCTION == "getER" {
							//

							condition.REQUEST = "RATE:CBR-RUSSIA:" + startRule.ARGUMENTS[0].VALUE

						} else if startRule.FUNCTION == "getSP" {
							//

							condition.REQUEST = "STOCK:ALPHAVANTAGE:" + startRule.ARGUMENTS[0].VALUE
						}

						allCallConditions[uniqueDapp.SIGNATURE] = condition
						// if _, ok := allDataSubscribes[condition.REQUEST]; !ok {
						// 	//
						//
						// 	allDataSubscribes[condition.REQUEST] = true
						// 	go subscriber.Subscribe("RATE", "CBR-RUSSIA", startRule.ARGUMENTS[0].VALUE, dataChannel, quitChannel)
						// }
						allDataSubscribes[condition.REQUEST] = true
						go subscriber.Subscribe("RATE", "CBR-RUSSIA", startRule.ARGUMENTS[0].VALUE, dataChannel, quitChannel)
					}
				}
			}

			var startstops []startstop
			val = result.Lookup("SMCOM").String()
			json.Unmarshal([]byte(val), &startstops)

			for _, uniqueStartStop := range startstops {
				//
				log.Println(uniqueStartStop)
				if uniqueStartStop.ACTION == "PAUSE" {
					//

					callCondition := allCallConditions[uniqueStartStop.SMS]
					callCondition.INWORK = false
					allCallConditions[uniqueStartStop.SMS] = callCondition

				} else if uniqueStartStop.ACTION == "START" {
					//

					callCondition := allCallConditions[uniqueStartStop.SMS]
					callCondition.INWORK = true
					allCallConditions[uniqueStartStop.SMS] = callCondition
				}
			}

			tools.MakeResponse(helpers.StatusOk, ctx)

		default:
			//

			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}

		return
	}

	ctx.Error("Unsupported method", fasthttp.StatusMethodNotAllowed)
}

func datareader() {
	//

	for {
		//

		select {

		case <-time.After(5 * time.Second):
			log.Println("TIMEOUT")

		case response := <-dataChannel:
			for key, callCondition := range allCallConditions {
				//

				log.Println(key, response.REQUEST)
				if callCondition.INWORK == false {
					//

					continue
				}

				if callCondition.REQUEST == response.REQUEST {
					//

					responseValue, err := strconv.ParseFloat(response.RESPOSNE["VALUE"], 10)
					if err != nil {
						//

						log.Println("FLOAT PARSE ERROR")
						continue
					}

					responseCount, err := strconv.ParseFloat(response.RESPOSNE["COUNT"], 10)
					if err != nil {
						//

						log.Println("FLOAT PARSE ERROR")
						continue
					}

					responseValue = responseValue / responseCount
					conditionValue, err := strconv.ParseFloat(callCondition.VALUE, 10)
					log.Println(responseValue, conditionValue, callCondition.CONDITION)
					if err != nil {
						//

						log.Println("FLOAT PARSE ERROR")
						continue
					}

					var ok bool = false

					if callCondition.CONDITION == "EQ" {
						//

						if conditionValue == responseValue {
							//

							ok = true
						}
					}

					if callCondition.CONDITION == "GT" {
						//

						if responseValue > conditionValue {
							//

							ok = true
						}
					}

					if callCondition.CONDITION == "GTE" {
						//

						if responseValue >= conditionValue {
							//

							ok = true
						}
					}

					if callCondition.CONDITION == "LT" {
						//

						if responseValue < conditionValue {
							//

							ok = true
						}
					}

					if callCondition.CONDITION == "LTE" {
						//

						if responseValue <= conditionValue {
							//

							ok = true
						}
					}

					if ok {
						//

						log.Println("MOTHER FUCKER!")

						var newCallForSign dappCallForSign
						var newCall dappCall
						newCall.TT, newCallForSign.TT = "SCC", "SCC"
						newCall.TST, newCallForSign.TST = strconv.FormatInt(time.Now().Unix(), 10), strconv.FormatInt(time.Now().Unix(), 10)
						newCall.ID, newCallForSign.ID = key, key
						newCall.SENDER, newCallForSign.SENDER = hex.EncodeToString(publicKey), hex.EncodeToString(publicKey)
						newCall.FUNC, newCallForSign.FUNC = callCondition.FUNC, callCondition.FUNC
						//newCall.ARGS = "SCC"
						//newCall.SIGNATURE = "SIGNATURE"

						newCallRaw, err := json.Marshal(newCallForSign)
						if err != nil {
							//

							log.Println("CAnt marshakk")
						}

						newCall.SIGNATURE = hex.EncodeToString(secp.Sign(newCallRaw, privateKey))
						newCallRaw, err = json.Marshal(newCall)
						if err != nil {
							//

							log.Println("CAnt marshakk")
						}

						_, err = http.Post("http://"+pvmHost+":25873/pvm/call", "application/json", bytes.NewBuffer(newCallRaw))
						if err != nil {
							//

							log.Println("CAnt POST", err)
						}
					}
				}
			}
		}
	}
}

func Start(m *mongo.Client, ip string, key []byte) {
	//

	var ok bool
	pvmHost, ok = os.LookupEnv("PVM_HOST_TCP_ADDR")
	if !ok {
		//

		pvmHost = "192.168.192.45"
	}

	mongoDB = m
	privateKey = key
	publicKey = secp.PubkeyFromSeckey(privateKey)

	server := &fasthttp.Server{
		Handler:          fastHTTPRawHandler,
		DisableKeepalive: true,
	}

	dataChannel = make(chan subscriber.Response, 1024)
	quitChannel = make(chan string)
	allCallConditions = make(map[string]callCondition)
	allDataSubscribes = make(map[string]bool)

	go datareader()

	panic(server.ListenAndServe(net.JoinHostPort(ip, "6060")))
}
