Welcome to our, The Fantastic Three's, Single Leader Replication program! This README will provide a step-to-step guide as to how to run the our program.

HOW TO RUN
    Step 1:
        To run the program, you'd first need to initialize three nodes and a client in four seperate terminals. The first three terminals you open are gonna be called node 0-2, written like this:
            go run . -id=x
        Replace x with the number node, so terminal 1 is id=0, terminal 2 is id=1 and terminal 3 is id=2. Be default, node 0 is the leader. 
    
    Step 2:
        For the fourth terminal, you run the following:
            go run ./client
        What the client does afterwards is ask for bidder ID and display the following options AFTER typing a bidder ID:
            1 — Bid: Place a bid.
            2 — Result: Check the highest bid and bidder.
            3 — Quit: Exit the client.
        And now the world's your oister (or however the saying goes)

COMMANDS
    Bid:
        Running this allows the user to place a bid in the auction. The user enters an amount, which is sent to the current leader. The results can be:
            success — The bid was accepted.
            fail — The bid was too low or the auction has ended.
            exception — No leader could be found after retrying.
    
    Result:
        Allows the user to check the current auction status. It displays the highest bid and the ID of the bidder who placed it. If the auction has ended, it displays the winning bidder and final amount.

    Quit:
        Closes the client application. The three server nodes continue running until they are stopped manually.
