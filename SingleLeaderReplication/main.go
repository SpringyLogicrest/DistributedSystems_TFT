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
	

}
