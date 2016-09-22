Tokens
======

Tokens are good for a single point in a single category.  They are
formed by prepending the category and a colon to the bubblebabble digest
of 3 random octets.  A token for the "merfing" category might look like
this:

    merfing:xunap-motex


Entropy
-------

3 octets provides 24 bits of entropy.  This gives 16777216 possible
tokens in each category.  The longest contest yet run lasted 24 hours,
which would give 2^24/24/60 = 11650 tokens per category per minute.  I
think this is a large enough pool to discourage brute-force attacks.
Assuming /dev/urandom is as good as is claimed, brute-force would be the
only way to attack it.


Token server
------------

Sometimes it's a good idea to have certain puzzles run on a different
machine than the server.  For instance, something that loads down the
CPU, or something that carries a high risk of local exploit.  The token
server listens on TCP port 1, issuing tokens encrypted with ARC4
(symmetric encryption).  Here's how the transaction goes:

    C: category
    S: nonce (4 bytes)
    C: nonce encrypted with symmetric key
    S: token encrypted with symmetric key


Token client
------------

The token client (in package "tokencli") runs as a daemon, requesting a
new token every minute for each puzzle.  Because we want you to have
multiple puzzles within a category, and the server only knows about
categories, each puzzle needs to be associated with a category.
Additionally, tokens are encrypted before being written to the local
filesystem, with a different key for each puzzle.

The token client thus needs a 4-tuple for each puzzle:

    (puzzle name, puzzle key, category, category key)

In the interest of making things easy to administer and code, this
4-tuple is stored in files and directories:

    /packages/packagename/tokencli/puzzle_name/enc.key
    /packages/packagename/tokencli/puzzle_name/category.key
    /packages/packagename/tokencli/puzzle_name/category

And puzzles are stored in:

    /state/tokens/puzzle_name

Using this scheme, the token client has only to iterate over
/packages/*/tokencli/* instead of implementing some sort of parser.
