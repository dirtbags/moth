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
