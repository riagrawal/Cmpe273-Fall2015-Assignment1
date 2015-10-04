// client.go
package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc/jsonrpc"
        "os"
        "strconv"
       
)

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

func main() {

	client, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	c := jsonrpc.NewClient(client)
	// Synchronous call
	if(len(os.Args)!=2){
		i, err := strconv.ParseFloat(os.Args[2], 32)
    		if err != nil {
        		log.Fatal("arith error:", err)
    		}
		args := &Args{os.Args[1],float32(i) }
        	reply := new(Result)
		err = c.Call("Calculator.Buy", args, reply)
		if err != nil {
			log.Fatal("arith error:", err)
		}
        	fmt.Printf("\n\nTradeId is : %d \n\nStocks: %s \n\nUnvestedAmount :$%.2f\n\n\n ",reply.Id,reply.Status,reply.Balance)	
	}else{ 
		id,_ := strconv.Atoi(os.Args[1])
		req := &TransId{id}
        	resp := new(Response)
		err = c.Call("Calculator.Response", req, resp)
		if err != nil {
			log.Fatal("Transaction Error :", err)
		}
        	fmt.Printf("\n\nStocks are : %s \n\nCurrentMarketValue is :$%.2f \n\nUnvestedAmount :$%.2f\n\n ",resp.Stocks,resp.CurrentMarketValue,resp.UnvestedAmount)	
	}
	
}