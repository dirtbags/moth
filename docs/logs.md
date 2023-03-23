Moth Logs
=======

Moth has multiple log channels: 

`points.log`
: the points log, used by server and scoreboard

`events.log`
: significant events, used to do manual analysis after an event

`stdout`
: HTTP server access 

`stderr`
: warnings and errors


`points.log` format
----------------------

The points log is a space-separated file.
Each line has four fields:

| `timestamp` | `teamID` | `category` | `points` |
| --- | --- | --- | --- |
| int | string | string | int |
| Unix epoch | Team's unique ID | Name of category | Points awarded |


### Example

```
1602702696 2255 nocode 1
1602702705 2255 sequence 1
1602702787 2255 nocode 2
1602702831 2255 sequence 2
1602702839 9458 nocode 3
1602702896 2255 sequence 8
1602702900 9458 nocode 4
1602702913 2255 sequence 16
```

`events.csv` format
----------------------

The events log is a comma-separated variable (CSV) file.
It ought to import into any spreadsheet program painlessly.

Each line has six fields minimum:

| `timestamp` | `event` | `teamID` | `category` | `points` | `extra`... |
| --- | --- | --- | --- | --- | --- |
| int | string | string | string | int | string... |
| Unix epoch | Event type | Team's unique ID | Name of category, if any | Points awarded, if any | Additional fields, if any |

Fields after `points` contain extra fields associated with the event. 

### Event types

These may change in the future.

* init: startup of server
* disabled: points accumulation disabled
* enabled: points accumulation re-enabled
* register: team registration
* load: puzzle load
* wrong: wrong answer submitted
* correct: correct answer submitted

### Example

```
1602716345,init,-,-,-,-,0
1602716349,load,2255,player5,sequence,1
1602716450,load,4824,player3,sequence,1
1602716359,correct,2255,player5,sequence,1
1602716423,wrong,4824,player3,sequence,1
1602716428,correct,4824,player3,sequence,1
1602716530,correct,4824,player3,sequence,1
1602716546,abduction,4824,player3,-,0,alien,FM1490
```

The final entry is a made-up "alien abduction" entry,
since at the time of writing,
we didn't have any actual events that wrote extra fields.
