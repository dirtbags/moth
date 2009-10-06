import xml.sax.saxutils

def mkDocTable(objects):
    objects.sort(lambda o1, o2: cmp(o1.__doc__, o2.__doc__))

    for object in objects:
        print '<table class="docs">'
        if object.__doc__ is None:
            print '<tr><th>%s<tr><td colspan=2>Bad object' % \
                  xml.sax.saxutils.escape(str(object))
            continue
        text = object.__doc__
        lines = text.split('\n')
        head = lines[0].strip()
        head = xml.sax.saxutils.escape(head)

        body = []
        for line in lines[1:]:
            line = line.strip() #xml.sax.saxutils.escape( line.strip() )
            line = line.replace('.', '.<BR>')
            body.append(line)

        body = '\n'.join(body)
        print '<DL><DT><DIV class="tab">%s</DIV></DT><DD>%s</DD></DL>' % (head, body)
        #print '<tr><th>%s<th>Intentionally blank<th><tr><td colspan=3>%s' % (head, body)
        print '</table>'

