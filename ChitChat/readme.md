Welcome to our, The Fantastic Three's, ChitChat program! This README provides step-by-step instructions on how to run our program, from starting the server and connecting clients to using the different chatroom features.

HOW TO RUN
    STEP 1:
        Once you have the project, open a new terminal that's inside the ChitChat folder (not in DistributedSystems_TFT_, but ChitChat). Let's call this "Terminal S".
        To check that the project compiles, you can run the following command: 
            go build ./...

    STEP 2:
        Run the following command in Terminal S: 
            go run ./server 
        This starts our ChitChat server on port 50051.
        You should see something similar to: 
            2026/10/23 22:31:36 Chat server started on port 50051 
        Of course, the date and time will vary. Keep Terminal S running while the upcoming clients are connected.

    STEP 3:
        Open a second terminal, Terminal A, and run: 
            go run ./client ClientA
        This starts ClientA, establishes a gRPC connection, subscribes to broadcast messages, and joins the chatroom.
        You should see something similar to:
            Welcome to ChitChat, ClientA!
            Connected to the server!
            [L:4] Participant ClientA joined Chit Chat at logical time 4
            You have joined ChitChat as ClientA!
        The logical timestamps may vary, and the join confirmation and broadcast can appear in either order.

    STEP 4:
        If you want to have multiple clients, you can open additional terminals inside the ChitChat folder. In this example, we'll do two.
        In Terminal B, run:
            go run ./client ClientB
        In Terminal C, run:
            go run ./client ClientC
        You should now have one server and three clients running simultaneously. Whenever a participant joins, all currently active participants receive a broadcast message, including the newly joined participant. For example, ClientA should eventually display:
            [L:4] Participant ClientA joined Chit Chat at logical time 4
            [L:7] Participant ClientB joined Chit Chat at logical time 7
            [L:10] Participant ClientC joined Chit Chat at logical time 10

MESSAGING
        Once a client has joined ChitChat, they can send messages directly from their terminal.
        Simply write whatever you want to say from a client terminal. Try for example:
            Hello everyone!
        The client sends the message to the server through the Publish RPC. The server then broadcasts it to connected participants with a Lamport logical timestamp.
        The server validates that messages are UTF-8 encoded and contain no more than 128 characters. Afterwards, all connected clients, including the sender, should receive the broadcast.

LEAVING
    To stop a client, press Ctrl+C in its terminal. This makes the server broadcast a departure notification to the remaining participants.
    For example:
        [L:23] Participant ClientC left Chit Chat at logical time 23

STOPPING THE SERVER
    Once all clients have finished, return to the server terminal and press Ctrl+C to stop the server.

LOGGING
    ChitChat uses Go's standard log library to record technical events. These logs can be found under the folder "logs".
    These logs contain subscription, join, and message-reception events, including logical timestamps. In addition, the client terminal displays more readable chat notifications instead of technical event logs, while the server currently writes its logs to the server terminal, so those should be saved when collecting evidence for the assignment report.