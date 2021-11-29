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
	
	"github.com/manifoldco/promptui"
)

var (
	id     int
	ports  []string
	ownBid int32
	Results []BidOrResult
)

type BidOrResult struct {
	Title string
}



func main() {
		f, err := os.OpenFile("../AuctionHouse Log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	Results = append(Results, BidOrResult{Title: "Bid"})
	Results = append(Results, BidOrResult{Title: "Result"})

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
	templates := promptui.SelectTemplates{
		Active:   `💰 {{ .Title | green | bold }}`,
		Selected: `{{ "✔" | green | bold }}`,
	}
	fmt.Println("-- Welcome to the AuctionHouse --\n\n")
	fmt.Println("-- Today you have a chance to get your hands on the brand new thing you want!!! --\n\n")
	fmt.Printf("Your ID is: %d \n\n", id)

	for {

		prompt := promptui.Select{
			Label: "Select",
			Items: Results,
			Templates: &templates,
		}
	
		_, result, err := prompt.Run()
	
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
	
		fmt.Printf("You selected %q\n", result)
		reader := bufio.NewReader(os.Stdin)
		if result == "{Bid}" {
			fmt.Println("ENTER YOUR BID 💰:")
			bidString, _ := reader.ReadString('\n')
			bidString = strings.TrimRight(bidString, "\r\n")
			bid, err := strconv.Atoi(bidString)

			if err != nil {
				fmt.Println("Wrong input!")
			} else if int32(bid) < ownBid {
				fmt.Println("You can't bid lower than your own previous bids! Try again.")
			} else {
				ownBid = int32(bid)
				Bid(int32(bid))
			}

		} else if result == "{Result}" {
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
	log.Printf("Response from server: Client{%d} bid {%d} -- %s\n", id, amount, bidResponse)

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
