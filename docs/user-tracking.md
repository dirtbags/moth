# User Tracking

We need some way to have track users uniquely.


## Motivation

### Individual progress

We're way too far gone on this one.
I fought it while I could,
but everybody and their dog wants to track individual progress,
so we need to continue providing at least advisory information about who's doing what.

### Attendance

CPE certificates are the biggest driver here.
Doing this client-side won't work,
because people want to fight me about their certificates,
and I need something to fall back on.

The sponsor also has a keen interest in attrition,
and we need attendance data for this as well.

### Chat

We need to integrate a chat system,
and for our big events,
we need the chat system to use the "display name" provided by each participant.


## Requirements

Essentially, we need something like team ID,
but for an individual participant.

### Support drop-in events

One of our big wins right now is our ability to run drop-in events,
like Def Con contests,
high school science cafes,
etc.

We dealt with this by pre-generating authentication tokens and providing a 
`/register` API endpoint to set a team name.
This was a good design and we should keep this.

### Run without Internet

Def Con's network is crap,
and we may yet run another event that's disconnected.
We need a way to run events without an Internet connection.

### Minimal storage

If possible, I'd prefer to not even have a password.
Ideally just a token for user, and their display name.


## Solution

I'm realizing the best solution is to do almost nothing.

We already have a client that provides a "participant ID",
which is logged into the event log.

The new chat system could pretty easily cache a mapping of `pid` to display name.
On cache miss, it could use whatever backend is provided to look things up.
This could be alfio, a URL to a CSV file, or something else.

