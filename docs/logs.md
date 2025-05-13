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


Awards log format
----------------------

The awards log is a comma-separated file recording when a team first correctly answers a puzzle.
Each line has four fields:

| `timestamp` | `teamID` | `category` | `puzzle` |
| --- | --- | --- | --- |
| int | string | string | int |
| Unix epoch | Team ID | Category ID | Puzzle ID |

### Example

```
1602702696,2255,nocode,1
1602702705,2255,sequence,1
1602702787,2255,nocode,2
1602702831,2255,sequence,2
1602702839,9458,nocode,3
1602702872,2255,heath,01-introduction
1602702896,2255,sequence,8
1602702900,9458,nocode,4
1602702910,2255,heath,02-discovery
1602702913,2255,sequence,16
```

Events log format
----------------------

The events log is a comma-separated variable (CSV) file.
It ought to import into any spreadsheet program painlessly.

Each line has six fields minimum:

| `timestamp` | `event` | `teamID` | `category` | `puzzle` | `extra`... |
| --- | --- | --- | --- | --- | --- |
| int | string | string | string | int | string... |
| Unix epoch | Event type | Team ID | Category ID, if any | Puzzle ID, if any | Additional fields, if any |

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
