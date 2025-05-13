Frequently Asked Questions
=================

## Why don't you implement HTTPS?

HTTPS can be handled more effectively by a front-end service such as Caddy, Traefik, or nginx.
Many HTTPS front-ends now automatically obtain and refresh an SSL certificate,
and provide other helpful features.

Due to the existence of said front-ends,
this question is a lot less frequent in 2025 than it was in 2010.
Everybody else gets it now ;)


## Why don't you implement IP Rate Limiting on the registration endpoint?

Any IP Rate Limiting solution the MOTH development team comes up with
would be inferior to the solutions provided by Caddy, Traefik, or nginx.
We're trying to make the best puzzle server possible,
other teams are focused on making the best IP Rate Limiter.

A better mechanism to prevent registration attacks is to provide 
a list of valid team IDs in `teamids.txt`.
There is substantially less need for an IP Rate Limiter if this facility can be used.
