a) What are packages in your implementation? What data structure do you use to transmit data and meta-data?
    In our implementation, packages are string messages. The messages “SYN” “SYN-ACK” and “ACK” represent the three packages that are exchanged during the TCP three-way handshake. We use channels of type chan string to the send/recieve messages between goroutines(client and server). Since our implementation mostly just simulates the three-way handshake, and is on the simpler side, we don’t use a separate datastructure for meta data.
b) Does your implementation use threads or processes? Why is it not realistic to use threads?
    In our implementation we use goroutines which represents threads. Here the client and server are different goroutines and communicate via channels. Using threads is not realistic because a real TCP normally communicates over a network and not through channels. 
c) In case the network changes the order in which messages are delivered, how would you handle message re-ordering?
    Our current implementation does not explicitly handle message reordering. One possible way to support it would be to attach sequence numbers to messages.
    The receiver could keep track of which sequence number it expects next. If, for example, message 2 arrived before message 1, message 2 could temporarily be stored in a buffer or priority queue. After message 1 arrived and was processed, the receiver could then process message 2.
d) In case messages can be delayed or lost, how does your implementation handle message loss?
    Our implementation handles message loss using timeouts and retransmissions. The forwarder introduces a random delay between 0 and 999 milliseconds and has a 20% chance of dropping each message.
    After sending a SYN, the client waits up to 2 seconds for a SYN-ACK. If it does not receive one, it sends another SYN. Similarly, after sending a SYN-ACK, the server waits up to 2 seconds for an ACK. If the ACK does not arrive, the server sends the SYN-ACK again.
    The server also handles receiving another SYN from the client. It assumes that its previous SYN-ACK may have been lost and sends another SYN-ACK. Likewise, after the client sends its ACK, it remains active for 3 seconds. If it receives another SYN-ACK during this period, it assumes that its previous ACK was lost and sends the ACK again.
    Therefore, delayed messages will normally still be processed, while lost messages cause retransmissions after a timeout.
e) Why is the 3-way handshake important?
    The three-way handshake is important because it helps verify that both the client and server are ready to communicate. 
