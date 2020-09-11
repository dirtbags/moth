---
pre:
  authors: 
    - neale
  attachments:
    - filename: salad.jpg
    - filename: s2.jpg
      filesystempath: salad2.jpg
debug:
  summary: Static puzzle resource files
answers: 
  - salad
---

You can include additional resources in a static puzzle,
by dropping them in the directory and listing them under `attachments`.

If the puzzle compiler sees both `filename` and `filesystempath`,
it changes the filename when the puzzle category is built.
You can use this to give good filenames while building,
but obscure them during build.
On this page, we obscure 
`salad2.jpg` to `s2.jpg`,
so that people can't guess the answer based on filename.

Check the source to this puzzle to see how this is done!

You can refer to resources directly in your Markdown,
or use them however else you see fit.
They will appear in the same directory on the web server once the exercise is running.
Check the source for this puzzle to see how it was created.

![Leafy Green Deliciousness](salad.jpg)
![Mmm so good](s2.jpg)

The answer for this page is what is featured in the photograph.
