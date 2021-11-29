package main

import (
	a "Auction/proto"
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	id    int
	ports []string
)

func main() {

	file, err := os.Open("../ports.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		ports = append(ports, scanner.Text())
	}

	RegisterClient()

	fmt.Println("-- Welcome to the AuctionHouse --\n\n")
	fmt.Println("-- Today you have a chance to get your hands on the brand new thing you want!!! --\n\n")
	fmt.Printf("Your ID is: %d \n\n", id)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("To bid press 1, To see the result press 2")
		text, _ := reader.ReadString('\n')
		text = strings.TrimRight(text, "\r\n")
		if text == "1" {
			fmt.Println("ENTER YOUR BID 💰:")
			bidString, _ := reader.ReadString('\n')
			bidString = strings.TrimRight(bidString, "\r\n")
			bid, err := strconv.Atoi(bidString)

			if err != nil {
				fmt.Println("Wrong input!")
			} else {
				Bid(int32(bid))
			}

		} else if text == "2" {
			fmt.Println("RESULT")
			Result()
		} else {
			fmt.Println("Wrong input!")
		}
		fmt.Println("\n\n")
	}

}

func RegisterClient() {
	for _, port := range ports {

		var conn *grpc.ClientConn
		conn, err := grpc.Dial(":"+port, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect to %s: %s", port, err)
		}

		defer conn.Close()
		AH := a.NewAuctionHouseClient(conn)

		response, err := AH.RegisterClient(context.Background(), &emptypb.Empty{})
		id = int(response.GetId())

		if err != nil {
			log.Fatalf("Error when registering Client: %s", err)
		}
		log.Printf("Response from port %s: %s\n", port, response.Acknowledgement)
		log.Printf("Your ID is: %d\n", id)
	}
}

func Bid(amount int32) {
	var bidResponse string = ""

	for _, port := range ports {

		var conn *grpc.ClientConn
		conn, err := grpc.Dial(":"+port, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect to %s: %s", port, err)
		}

		defer conn.Close()
		AH := a.NewAuctionHouseClient(conn)

		response, err := AH.Bid(context.Background(), &a.Request{Id: int32(id), Amount: amount})

		if err != nil {
			log.Printf("(THE CLIENT WOULD NOT SEE THIS) ------ Error when calling Bid: A server crashed!")
		} else {
			bidResponse = response.Acknowledgement
		}
	}
	log.Printf("Response from server: %s\n", bidResponse)

}

func Result() {

	var highestCurrentBid int32 = 0
	var highestBidderID int32 = 0
	var isAuctionOver bool

	for _, port := range ports {

		var conn *grpc.ClientConn
		conn, err := grpc.Dial(":"+port, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect to %s: %s", port, err)
		}

		defer conn.Close()
		AH := a.NewAuctionHouseClient(conn)

		outcome, err := AH.Result(context.Background(), &emptypb.Empty{})

		if err != nil {
			log.Printf("(THE CLIENT WOULD NOT SEE THIS) ------ Error when calling Result: A server crashed!")
		} else {

			if outcome.GetIsOver() {
				isAuctionOver = true
			}

			if outcome.HighestBid > highestCurrentBid {
				highestCurrentBid = outcome.GetHighestBid()
				highestBidderID = outcome.GetId()
			}
		}

	}

	if isAuctionOver {
		if highestBidderID != int32(id) {
			log.Printf("Auction is over - Winner is client %d with the bid of %d 💰\n", highestBidderID, highestCurrentBid)
		} else {
			log.Printf("Auction is over - Winner is YOU with the bid of %d 💰\n", highestCurrentBid)
		}
	} else {
		log.Printf("The Auction is still going, highest current bid = %d 💰\n", highestCurrentBid)
	}

}
