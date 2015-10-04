// server.go
package main

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
        	"strings"
        	"strconv"
	"time"
	"math/rand"
	"net/http"
	"io/ioutil"
	"encoding/json"	
	"fmt"
)

const (
	timeout = time.Duration(time.Second * 10)
)

type StockDetail struct {
	List struct {
		Resources []struct {
			Resource struct {
				Fields struct {
					Name    string `json:"name"`
					Price   string `json:"price"`
					Symbol  string `json:"symbol"`
					Ts      string `json:"ts"`
					Type    string `json:"type"`
					UTCTime string `json:"utctime"`
					Volume  string `json:"volume"`
				} `json:"fields"`
			} `json:"resource"`
		} `json:"resources"`
	} `json:"list"`
}


type Args struct {
	Info string
        Budget float32
}

type TransId struct {
	Tid int
}

type Result struct{
 
Id int
Status string
Balance float32

}

type Response struct {

Stocks string
CurrentMarketValue float32
UnvestedAmount float32

}

type Stock struct{

symbol string
balance float32
num int 
price float64 

}

type StockResponse struct{

symbol string
curMarketValue float32
num int 
newPrice float64 


}

func stockToString(stock *Stock) string{
	return "\""+ stock.symbol + ":" + strconv.Itoa(stock.num) + ":$" + strconv.FormatFloat(stock.price, 'f', 2, 32) + "\""
}

func returnStocksstring (stock []Stock) string{
	var str string = ""
	for j := range stock {
 	str = str  + stockToString(&stock[j]) + ","
	}
	str = str[:len(str)-1]
	return str
	
}

func stockResponseToString(stock *StockResponse) string{
	return "\""+ stock.symbol + ":" + strconv.Itoa(stock.num) + ":$" + strconv.FormatFloat(stock.newPrice, 'f', 2, 32) + "\""
}

func returnStocksResponseString (stock []StockResponse) string{
	var str string = ""
	for j := range stock {
 	str = str  + stockResponseToString(&stock[j]) + ","
	}
	str = str[:len(str)-1]
	return str
	
}


type Calculator struct{}

var m map[int] []Stock
var keys [] int
var bget map[int] float32
var keyb [] int

func (t *Calculator) Buy(args *Args, reply *Result) error {
	
	stockList := Returnlistofstocks(args)
	rand.Seed(time.Now().UTC().UnixNano())   
        reply.Id = rand.Int()
        reply.Status = returnStocksstring (stockList)
        reply.Balance = getRemainingBalance(stockList,args.Budget)
	m[reply.Id] = stockList
	bget[reply.Id] = args.Budget
	return nil
}

func (t *Calculator) Response(args *TransId, resp *Response) error {
	
	if(m[args.Tid]!=nil){
	stocklist := ReturnlistofResponseStocks(m[args.Tid]) 
        resp.Stocks = returnStocksResponseString (stocklist)
        resp.CurrentMarketValue = getCurrentValue(stocklist)
	resp.UnvestedAmount = getRemainingBalance(m[args.Tid],bget[args.Tid])
	return nil
	}else{
	err := fmt.Errorf("user (id %d) not found",args.Tid)
	return err
	}
}

func main() {
	m = make(map[int] []Stock)
	bget = make(map[int] float32)
	cal := new(Calculator)
	server := rpc.NewServer()
	server.Register(cal)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			log.Printf("new connection established\n")
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	
	}
}


func Returnlistofstocks (args * Args) []Stock{
		
	data :=strings.Split(args.Info,",")
	stockList := []Stock{}
	for i := range data {
             no:=strings.Split(data[i],":")
	     var stock Stock   // create variable
             stock.symbol = (no[0])
             n:=strings.Split(no[1],"%")
	     var percent int
             percent,_= strconv.Atoi(n[0])
	     stock.balance = (float32(percent)/float32(100))*float32(args.Budget)
	     st,error := GetQuote(stock.symbol)
	     if error!=nil{
			fmt.Println(error)
	     }
	     var pr float64
	     pr,_ = st.GetPrice()
	     stock.price = pr
	     stock.num = int(stock.balance/float32(pr))
	     var difference float32 = (stock.balance/float32(pr))- float32(int(stock.balance/float32(pr)))
	     stock.balance = difference * float32(pr)
	     stockList = append(stockList, stock)
	   
        }
	return stockList
	
}

func ReturnlistofResponseStocks (stockList []Stock) []StockResponse{
	stocklist:= []StockResponse{}
	for k:= range stockList{
		var responseStock StockResponse
		responseStock.symbol = stockList[k].symbol
		responseStock.num = stockList[k].num
		st,error := GetQuote(stockList[k].symbol)
		if error!=nil{
			fmt.Println(error)
	     	}
		var pr float64
	     	pr,_ = st.GetPrice()
		if(pr >= stockList[k].price){
		responseStock.newPrice = pr
		}else{
		responseStock.newPrice = -pr
		}
		responseStock.curMarketValue = float32(pr*float64(stockList[k].num) )
		stocklist = append(stocklist, responseStock)
	}	
	return stocklist
}


func getRemainingBalance (stockList []Stock,budget float32) float32 {
	var totalValue float32 = 0.0
	for k:= range stockList	 {
		totalValue = totalValue + (float32(stockList[k].num) * float32(stockList[k].price))
        }
	return (float32(budget)- totalValue)
}

func getCurrentValue (stockList []StockResponse) float32 {
	var currentValue float32 = 0.0
	for k:= range stockList	 {
		currentValue = currentValue + stockList[k].curMarketValue
        }
	return currentValue
}



// get full stock details into a struct
func GetQuote(symbol string) (StockDetail, error) {
	// set http client timeout
	client := http.Client{Timeout: timeout}

	url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json", symbol)
	res, err := client.Get(url)
	if err != nil {
		return StockDetail{}, fmt.Errorf("Stocks cannot access yahoo finance API: %v", err)
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return StockDetail{}, fmt.Errorf("Stocks cannot read json body: %v", err)
	}

	var stock StockDetail

	err = json.Unmarshal(content, &stock)
	if err != nil {
		return StockDetail{}, fmt.Errorf("Stocks cannot parse json data: %v", err)
	}
	return stock, nil
}


// return the stock price
func (stock StockDetail) GetPrice() (float64, error) {
	price, err := strconv.ParseFloat(stock.List.Resources[0].Resource.Fields.Price, 64)
	if err != nil {
		return 1.0, fmt.Errorf("Stock price: %v", err)
	}
	return price, nil
}

func (stock StockDetail) String() float64 {
	price, err := stock.GetPrice()
	if err != nil {
		fmt.Printf("Error getting price: %v", err)
	}
	return price
}
