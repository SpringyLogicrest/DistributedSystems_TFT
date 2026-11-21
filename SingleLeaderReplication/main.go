package main
 
import (
	"flag"
	"log"
	"net"
	"time"
 
	pb "SingleLeaderReplication/proto"
 
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)
 
// The three nodes. Node i listens on addrs[i].

var addrs = []string{"localhost:5000", "localhost:5001", "localhost:5002"}


func main() {
	// Get the node's ID from the terminal arguments.
	id := flag.Int("id", 0, "Node ID")
	end := flag.Int("end", 0, "End node ID")
	flag.Parse()

	n := &Node{
		ID:      *id,
		endTime:  time.Unix(*end, 0),
		bidders:  map[string] bool{},

	}

	for _, a := range addrs {
		conn, err := grpc.NewClient(a, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		n.conns = append(n.conns, conn)
	}


	lis, err := net.Listen("tcp", addrs[*id])
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer()
	pb.RegisterAuctionServer(server, n)
	pb.RegisterReplicationServer(server, n)

	log.Printf("node %d listening on %s", *id, addrs[*id])
	log.Fatal(server.Serve(lis))



}
