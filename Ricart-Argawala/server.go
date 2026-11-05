//recieves and handles grpc messages

package ricartargawala

import (
	"context"
	"fmt"

	pb "ricart-agrawala/proto"
)

type MutexServer struct {
	pb.UnimplementedMutexServiceServer
	node *Node
}

func (s *MutexServer) RequestAccess(
	ctx context.Context,
	req *pb.Request,
) (*pb.Reply, error) {

	fmt.Printf("Node %d recieved request from Node %d", s.node.ID, req.NodeId)
	return &pb.Reply{Granted: true}, nil
}
