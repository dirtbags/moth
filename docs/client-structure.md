MOTH Client Directory Structure
=======

MOTHv5 implements WebDAV.
Depending on the authentication level of a user,
files may be read-only, or read-write.

WebDAV allows you to mount MOTH as a local filesystem.
You are encouraged to do this,
and use this document as a reference.


Directory Structure: Participant
-----------

Here is an example list of the files available to participants.

    r- /state/points.log
    rw /state/self/name
    rw /state/self/private.dat
    rw /state/self/public.dat
    r- /state/1/name
    r- /state/1/public.dat
    r- /state/2/name
    r- /state/2/public.dat
    r- /puzzles/category-a/1/index.html
    rw /puzzles/category-a/1/answer
    r- /puzzles/category-a/2/index.html
    rw /puzzles/category-a/2/answer
    r- /puzzles/category-a/2/attachment.jpg
    r- /puzzles/category-b/1/index.html
    rw /puzzles/category-b/1/answer
    r- /theme/*

Directory Structure: Anonymous
-----------

Anonymous (unauthenticated) users
have a restricted view:

    r- /state/points.log
    rw /state/self/name
    rw /state/self/private.dat
    rw /state/self/public.dat
    r- /state/1/name
    r- /state/1/public.dat
    r- /state/2/name
    r- /state/2/public.dat


Directory Structure: Administrator
------------

Here is an example list of the files available
to an administrator.
