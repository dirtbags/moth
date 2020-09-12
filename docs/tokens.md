Tokens
======

We used to use tokens extensively for categories outside of MOTH
(like scavenger hunts, Dirtbags Tanks, and other standalone stuff).

We still occasionally pull out tokens to deal with oddball categories
that we want to score alongside MOTH categories.

Here's how they work.

Description
------------

Tokens are a 3-tuple:

> (category, points, nonce)

We build a mothball with nothing but `answers.txt`,
and a special 1-point puzzle that uses JavaScript to parse and submit tokens.

Generally, tokens use colon separators, so they look like this:

    category:12:xunap-motex

Uniqueness
--------

Because they work just like normal categories,
you can't have two distinct tokens worth the same number of points.

When we need two or more tokens worth the same amount,
we make the point values very high,
so the least significant digit doesn't have much impact on the overall value.
For instance:

    category:1000001:xylep-nanox
    category:1000002:xenod-relix
    category:1000003:xoter-darox


Entropy
-------

3 octets provides 24 bits of entropy.  This gives 16777216 possible
tokens in each category.  The longest contest yet run lasted 24 hours,
which would give 2^24/24/60 = 11650 tokens per category per minute.  I
think this is a large enough pool to discourage brute-force attacks.
Assuming /dev/urandom is as good as is claimed, brute-force would be the
only way to attack it.
