Philosophy
==========

This is just some scattered thoughts by the architect, Neale.

People are going to try to break this thing.
It needs to be bulletproof.
This pretty much set the entire design:

* As much as possible is done client-side
  * Participants can attack their own web browsers as much as they feel like
  * Also reduces server load
  * We will help you create brute-force attacks!
    * Your laptop is faster than our server
    * We give you the carrot of hashed answers and the hashing function
    * This removes one incentive to DoS the server
* Generate static content whenever possible
  * Puzzles are statically compiled before the event even starts
  * `points.json` and `puzzles.json` are generated and cached by a maintenance loop
* Minimize dynamic handling
  * There are only two (2) dynamic handlers
    * team registration
    * answer validation
  * You can disable team registration if you want, just remove `teamids.txt`
  * I even removed token handling once I realized we replicate the user experience with the `answer` handler and some client-side JavaScript
* As much as possible is read-only
  * The only rw directory is `state`
* Server code should be as tiny as possible
  * Server should provide highly limited functionality
  * It should be easy to remember in your head everything it does
* Server is also compiled
  * Static type-checking helps assure no run-time errors
* Server only tracks who scored how many points at what time
  * This means the scoreboard program determines rankings
  * Want to provide a time bonus for quick answers? I don't, but if you do, you can just modify the scoreboard to do so.
  * Maybe you want to show a graph of team rankings over time: just replay the event log.
  * Want to do some analysis of what puzzles take the longest to answer? It's all there.
