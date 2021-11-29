package main

import (
	a "Auction/proto"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
	"github.com/manifoldco/promptui"
	"strings"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ReplicaManager struct {
	a.UnimplementedAuctionHouseServer

	//ID              int
	port            string
	highestBid      int32
	highestBidderID int32
	Bidders         []int
	AllClients      []int
	Clients         int32
	isOver          bool
}

var (
	ReplicaManagers []ReplicaManager
	//Port            string
	auctionOver     bool
	Ports           []Port
)

type Port struct {
	PortNumber string
}

func main() {
	f, err := os.OpenFile("../AuctionHouse Log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	//Adds port number to be displayed in the UI
	Ports = append(Ports, Port{PortNumber: "5000"})
	Ports = append(Ports, Port{PortNumber: "5001"})
	Ports = append(Ports, Port{PortNumber: "5002"})
	//To add more, simply reuse the above and add a new port number

	//Customizing the UI
	templates := promptui.SelectTemplates{
		Active:   `📶 {{ .PortNumber | green | bold }}`,
		Selected: `{{ "✔" | green | bold }}`,
	}

	//The UI Prompt itself where "Items" is which "set/list" to display
	prompt := promptui.Select{
		Label: "Port Number",
		Items: Ports,
		Templates: &templates,
	}

	//Runs the UI which returns the selected Item
	_, result, err := prompt.Run()
	
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	//Since the returned item is from a list it is returned as {Item} and we need to remove {}
	result = strings.Replace(result, "{", "", -1)
	result = strings.Replace(result, "}", "", -1)
	auctionOver = true
	server := ReplicaManager{
		port: result,
	}
	go StartServer(result, server)
	fmt.Println("Started server with port: " + result)

	/*fmt.Println("Enter port number (You can only choose between 5000, 5001 and 5002): ")
	reader := bufio.NewReader(os.Stdin)
	inputPort, _ := reader.ReadString('\n')
	inputPort = strings.TrimRight(inputPort, "\r\n")

	file, err := os.Open("../ports.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		scannedPort := scanner.Text()
		if scannedPort == inputPort {
			Port = scannedPort
			server := ReplicaManager{
				port: Port,
			}
			go StartServer(Port, server)
			fmt.Println("Started server with port: " + Port)
			break
		}
	}
	*/

	for {
		// Infinity loop
	}
}

func (AH *ReplicaManager) Bid(ctx context.Context, bid *a.Request) (*a.Response, error) {
	//If the auction is running register the bidder then check if bid is higher than highestCurrentBid
	if !auctionOver {

		if !contains(AH.Bidders, int(bid.Id)) {
			AH.Bidders = append(AH.Bidders, int(bid.Id))
		}
		if bid.GetAmount() > AH.highestBid {
			AH.highestBid = bid.GetAmount()
			AH.highestBidderID = bid.GetId()
			return &a.Response{Acknowledgement: "Your bid was registered"}, nil
		} else {
			return &a.Response{Acknowledgement: "Your bid was lower than the current bid, please check the outcome"}, nil
		}

		//If no auction is active, start a new one
	} else {
		auctionOver = false
		go AH.startAuction(time.Duration(30))
		AH.Bidders = append(AH.Bidders, int(bid.Id))
		AH.highestBid = bid.GetAmount()
		AH.highestBidderID = bid.GetId()

		bidAmount := strconv.FormatInt(int64(bid.GetAmount()), 10)

		return &a.Response{Acknowledgement: "No auction was running, you have started one with a bid of " + bidAmount}, nil
	}

}

func (AH *ReplicaManager) Result(ctx context.Context, _ *emptypb.Empty) (*a.Outcome, error) {
	if auctionOver {
		resultRespons := a.Outcome{
			Id:         AH.highestBidderID,
			HighestBid: AH.highestBid,
			IsOver:     true,
			Winner:     AH.highestBidderID,
		}
		return &resultRespons, nil
	} else {
		resultRespons := a.Outcome{
			Id:         AH.highestBidderID,
			HighestBid: AH.highestBid,
			IsOver:     AH.isOver,
			Winner:     AH.highestBidderID,
		}
		return &resultRespons, nil
	}
}

func (AH *ReplicaManager) RegisterClient(ctx context.Context, _ *emptypb.Empty) (*a.Response, error) {
	registerResponse := a.Response{
		Id:              AH.Clients,
		Acknowledgement: "Registered",
	}
	AH.AllClients = append(AH.AllClients, int(AH.Clients))
	AH.Clients++
	return &registerResponse, nil
}

func StartServer(port string, toRegister ReplicaManager) {
	list, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	a.RegisterAuctionHouseServer(grpcServer, &toRegister)
	err = grpcServer.Serve(list)
	if err != nil {
		log.Printf("Failed to start gRPC server: %v", err)
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
	log.Println("Auction started! The duration of the auction is ", timeInSec.String())
	time.Sleep(timeInSec * time.Second)
	auctionOver = true
	AH.Bidders = nil
	log.Println("AUCTION IS OVER!")
	outcome, err := AH.Result(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Printf("(THE CLIENT WOULD NOT SEE THIS) ------ Error when calling Result: A server crashed!")
	} else {
		log.Printf("Auction is over - Winner is client %d with the bid of %d 💰\n", outcome.Id, outcome.HighestBid)

	}


}
