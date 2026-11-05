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
        By typing 'probe', it currently sends a test request to the other nodes using gRPC. The receiving nodes respond, and the sending node prints the replies. Hereøs an example outpout:
            probe
            2026/11/05 13:07:56 Node 1: sending REQUEST to Node 2 (timestamp=1)
            2026/11/05 13:07:56 Node 1: received REPLY from Node 2 (granted=true)
            2026/11/05 13:07:56 Node 1: sending REQUEST to Node 3 (timestamp=3)
            2026/11/05 13:07:56 Node 1: received REPLY from Node 3 (granted=true)
    
    quit:
        By typing 'stop', it stops the node's gRPC server and exits the program.