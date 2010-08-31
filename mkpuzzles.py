#! /usr/bin/python

import os
import shutil
import optparse
import string
import markdown
import rfc822
from codecs import open

p = optparse.OptionParser()
p.add_option('-t', '--template', dest='template', default='template.html',
             help='Location of HTML template')
p.add_option('-b', '--base', dest='base', default='',
             help='Base URL for contest')
p.add_option('-p', '--puzzles', dest='puzzles', default='puzzles',
             help='Directory containing puzzles')
p.add_option('-w', '--htmldir', dest='htmldir', default='puzzler',
             help='Directory to write HTML puzzle tree')
p.add_option('-k', '--keyfile', dest='keyfile', default='puzzler.keys',
             help='Where to write keys')

opts, args = p.parse_args()

keys = []

tmpl_f = open(opts.template, encoding='utf-8')
template = string.Template(tmpl_f.read())
tmpl_f.close()

js = '''
<script type="text/javascript">
    function readCookie(key) {
        var s = key + '=';
        var toks = document.cookie.split(';');
        for (var i = 0; i < toks.length; i++) {
            var tok = toks[i];
            while (tok.charAt(0) == ' ') {
                tok = tok.substring(1, tok.length);
            }
            if (tok.indexOf(s) == 0) {
                return tok.substring(s.length, tok.length);
            }
        }
        return null;
    }

    function getTeamInfo() {
        team = readCookie('team');
        passwd = readCookie('passwd');
        if (team != null) {
            document.getElementById("form").t.value = team;
        }
        if (passwd != null) {
            document.getElementById("form").w.value = passwd;
        }
    }
</script>
'''

for cat in os.listdir(opts.puzzles):
    dirname = os.path.join(opts.puzzles, cat)
    for points in os.listdir(dirname):
        pointsdir = os.path.join(dirname, points)
        if not os.path.isdir(pointsdir):
            continue

        outdir = os.path.join(opts.htmldir, cat, points)
        try:
            os.makedirs(outdir)
        except OSError:
            pass

        readme = ''
        files = []
        for fn in os.listdir(pointsdir):
            path = os.path.join(pointsdir, fn)
            if fn == 'key':
                for key in open(path, encoding='utf-8'):
                    key = key.rstrip()
                    keys.append((cat, points, key))
            elif fn == 'hint':
                pass
            elif fn == 'index.exe':
                p = os.popen(path)
                m = rfc822.Message(p)
                for key in m.getallmatchingheaders('Key'):
                    print key
                    keys.append((cat, points, key))
                readme = m.fp.read()
                if m.get('Content-Type', 'text/markdown') == 'text/markdown':
                    readme = markdown.markdown(readme)
            elif fn == 'index.html':
                readme = open(path, encoding='utf-8').read()
            elif fn == 'index.mdwn':
                readme = open(path, encoding='utf-8').read()
                readme = markdown.markdown(readme)
            elif fn.endswith('~'):
                pass
            else:
                files.append((fn, path))

        title = '%s for %s points' % (cat, points)

        body = []
        if readme:
            body.append('<div class="readme">%s</div>\n' % readme)
        if files:
            body.append('<ul>\n')
            for fn, path in files:
                if os.path.isdir(path):
                    shutil.rmtree(os.path.join(outdir, fn), ignore_errors=True)
                    shutil.copytree(path, os.path.join(outdir, fn))
                else:
                    shutil.copy(path, outdir)

                if not fn.startswith(','):
                    body.append('<li><a href="%s">%s</a></li>\n' % (fn, fn))
            body.append('</ul>\n')
        body.append('''
    <form id="form" action="%(base)spuzzler.cgi" method="post">
      <fieldset>
        <legend>Your answer:</legend>
        <input type="hidden" name="c" value="%(cat)s" />
        <input type="hidden" name="p" value="%(points)s" />
        Team: <input name="t" /><br />
        Password: <input type="password" name="w" /><br />
        Key: <input name="k" /><br />
        <input type="submit" />
      </fieldset>
    </form>
''' % {'base': opts.base,
       'cat': cat,
       'points': points})

        page = template.substitute(hdr=js,
                                   title=title,
                                   base=opts.base,
                                   links='',
                                   body_class='',
                                   onload = "getTeamInfo()",
                                   body=''.join(body))

        f = open(os.path.join(outdir, 'index.html'), 'w', encoding='utf-8')
        f.write(page)

f = open(opts.keyfile, 'w', encoding='utf-8')
for key in keys:
    f.write('%s\t%s\t%s\n' % key)

