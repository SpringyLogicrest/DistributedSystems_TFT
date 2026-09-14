package wayhandshake
//3 way tcp handshake
import ("fmt",
"sync"
);




func server(syn <- chan string, ack <- chan string, synack <- chan string){
	// this tells the waitgroup that the server goroutine is done when it returns
	// the defer makes it like always run even if there is a error.
	defer wg.Done()
	msg := <- syn

	fmt.Println("Server has got the message: ", msg)
	
}

func client(syn <- chan string, ack <- chan string, synack <- chan string){
	defer wg.Done()
	syn <- "SYN"
	fmt.Println("Client has sent the message: SYN")
}



func main() {

	
}
	
