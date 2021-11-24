package main

import (
	a "Auction/proto"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ReplicaManager struct {
	a.UnimplementedAuctionHouseServer

	ID              int
	port            string
	highestBid      int32
	highestBidderID int32
	Bidders         []int
	Clients         int32
	isOver          bool
}

var (
	ReplicaManagers []ReplicaManager
)

func main() {

	file, err := os.Open("../ports.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		RM := ReplicaManager{
			ID:         len(ReplicaManagers) + 1,
			port:       scanner.Text(),
			highestBid: 0,
			Clients:    1,
		}
		ReplicaManagers = append(ReplicaManagers, RM)
	}
	for _, RM := range ReplicaManagers {
		go StartServer(RM.port, RM)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		//After a timeframe close the auction
		fmt.Println("ENTER SECONDS OF AUCTION PLS")
		auctionData, _ := reader.ReadString('\n')
		auctionData = strings.TrimRight(auctionData, "\r\n")
		fmt.Println("REGISTERED YOUR SECONDS", auctionData)
		interval, err := strconv.Atoi(auctionData)
		if err != nil {
		}
		for _, server := range ReplicaManagers {
			go server.startAuction(time.Duration(interval))
		}
	}

}

//Missing return for exception (potentially not needed)
func (AH *ReplicaManager) Bid(ctx context.Context, bid *a.Request) (*a.Response, error) {
	if !contains(AH.Bidders, int(bid.Id)) {
		AH.Bidders = append(AH.Bidders, int(bid.Id))
	}

	if bid.GetAmount() > AH.highestBid {
		AH.highestBid = bid.GetAmount()
		AH.highestBidderID = bid.GetId()
		return &a.Response{Acknowledgement: "Your bid was registered"}, nil
	} else {
		return &a.Response{Acknowledgement: "Your bid was too low, please check the outcome"}, nil
	}

}

func (AH *ReplicaManager) Result(ctx context.Context, _ *emptypb.Empty) (*a.Outcome, error) {
	fmt.Println("Sending back isOver", AH.isOver)
	resultRespons := a.Outcome{
		Id:         AH.highestBidderID,
		HighestBid: AH.highestBid,
		IsOver:     AH.isOver,
		Winner:     AH.highestBidderID,
	}
	return &resultRespons, nil
}

func (AH *ReplicaManager) RegisterClient(ctx context.Context, _ *emptypb.Empty) (*a.Response, error) {
	registerResponse := a.Response{
		Id:              AH.Clients,
		Acknowledgement: "Registered",
	}
	AH.Clients++
	return &registerResponse, nil
}

func StartServer(port string, toRegister ReplicaManager) {
	list, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	a.RegisterAuctionHouseServer(grpcServer, &toRegister)
	err = grpcServer.Serve(list)
	if err != nil {
		fmt.Printf("Failed to start gRPC server: %v", err)
	}
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (AH *ReplicaManager) startAuction(timeInSec time.Duration) {
	fmt.Println("Auction started!")
	AH.isOver = false
	time.Sleep(timeInSec * time.Second)
	AH.isOver = true

	fmt.Println("AUCTION IS OVER!")
	fmt.Println(AH.isOver)
}
