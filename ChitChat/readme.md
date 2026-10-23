Welcome to our, The Fantastic Three's, ChitChat program! This document will contain a step by step instruction on how to run our program, from starting it up to doing various things in the chatrooms.

STEP 1:
    Once you have the project, open a new terminal that's inside the ChitChat folder. We'll call this "Terminal A".

STEP 2:
    Run the following command in terminal A: "go run ./server". This will start the server for our program.
    You should get the following message: "2026/10/23 20:13:13 Chat server started on port 50051".

STEP 3:
    Open a second terminal, "Terminal B", and run the following command in terminal B "go run ./client". This will connect the client to the server we started previously.
    You should get the following message in terminal B: "2026/10/23 20:14:43 Joined chatroom as ClientA with logical timestamp 4".
    Meanwhile, in terminal A, you should get the following message: "2026/10/23 20:14:43 ClientA  joined the chatroom! logical time  3"

These are the basics steps we have for starting the server and connecting a client. With that, below will be the different commands you can do as a client:
    TO BE ADDED