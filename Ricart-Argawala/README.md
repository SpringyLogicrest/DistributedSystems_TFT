Welcome to our, The Fantastic Three's, Ricart-Argawala program! This README will provide a step-to-step guide as to how to run the our program.

HOW TO RUN
    Step 1:
        If we want to run our program, you'd need initialize all three possible nodes on a terminal of their own. Before we do that, we need to build the program, done with the following command in an integrated terminals to the root folder 'Ricart-Argawala':
            go build .

    Step 2:
        Now that the program has been built, open three integrated terminals to the root folder 'Ricart-Argawala'.

    Step 3:
        In the first terminal, write the following command:
            go run . -id=1
        This should give you the following:
            Commands: probe, quit
            > 2026/11/05 13:07:17 Node 1: listening on localhost:50051
        The date and time will of course vary depending on your PC's date and time.

    Step 4:
        Similarly to what you did in step 3, go to each of the two other integrated terminals and write in the second terminal:
            go run . -id=2
        And in the third terminal:
            go run . -id=3
        This should produce similar results as to that of step 3, but with different localhosts incrimented by one digit.

COMMANDS
    probe:
        By typing 'probe', we requests access to the Critical Section using the Ricart–Agrawala algorithm. The node sends requests to the other nodes and waits until it receives permission from all of them. It then enters the Critical Section for 10 seconds before exiting.
            probe
            Node 1: requesting access to critical section (timestamp=1)
            2026/11/06 15:15:44 Node 1: sending REQUEST to Node 2 (timestamp=1)
            2026/11/06 15:15:44 Node 1: sending REQUEST to Node 3 (timestamp=1)
            2026/11/06 15:15:44 Node 1: received REPLY from Node 2 (granted=true)
            2026/11/06 15:15:44 Node 1: received REPLY from Node 3 (granted=true)
            Node 1: entering critical section
        If no other probe requests have been made from other nodes while the current node is in the critical section, it'll wait 10 seconds before declaring:
            Node 1: is going to exit the critical section
        Here's an example where Node 2 runs the probe command and recieves a request from Node 3 while in the critical section:
            probe
            Node 2: requesting access to critical section (timestamp=3)
            2026/11/06 15:17:01 Node 2: sending REQUEST to Node 3 (timestamp=3)
            2026/11/06 15:17:01 Node 2: sending REQUEST to Node 1 (timestamp=3)
            2026/11/06 15:17:01 Node 2: received REPLY from Node 3 (granted=true)
            2026/11/06 15:17:01 Node 2: received REPLY from Node 1 (granted=true)
            Node 2: entering critical section
            2026/11/06 15:17:04 Node 2: received REQUEST from Node 3 (timestamp=5, clock=6)
            2026/11/06 15:17:04 Node 2: deffering reply to Node 3
            Node 2: is going to exit the critical section
            > 2026/11/06 15:17:11 Node 2: sending REPLY to Node 3
        Here's what it looks like for Node 3:
            probe
            Node 3: requesting access to critical section (timestamp=5)
            2026/11/06 15:17:04 Node 3: sending REQUEST to Node 1 (timestamp=5)
            2026/11/06 15:17:04 Node 3: sending REQUEST to Node 2 (timestamp=5)
            2026/11/06 15:17:04 Node 3: received REPLY from Node 1 (granted=true)
            2026/11/06 15:17:11 Node 3: received REPLY from Node 2 (granted=true)
            Node 3: entering critical section
            Node 3: is going to exit the critical section

    quit:
        By typing 'quit', it stops the node's gRPC server and exits the program.