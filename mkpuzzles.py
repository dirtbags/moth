#! /usr/bin/env python3

import os
import shutil
import optparse
from ctf import config

p = optparse.OptionParser()
p.add_option('-p', '--puzzles', dest='puzzles', default='puzzles',
             help='Directory containing puzzles')
p.add_option('-w', '--htmldir', dest='htmldir', default='puzzler',
             help='Directory to write HTML puzzle tree')
p.add_option('-k', '--keyfile', dest='keyfile', default='puzzler.keys',
             help='Where to write keys')

opts, args = p.parse_args()

keys = []

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
    window.onload = getTeamInfo;
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
                key = open(path, encoding='utf-8').readline().strip()
                keys.append((cat, points, key))
            elif fn == 'index.html':
                readme = open(path, encoding='utf-8').read()
            elif fn.endswith('~'):
                pass
            else:
                files.append((fn, path))

        title = '%s for %s points' % (cat, points)
        f = open(os.path.join(outdir, 'index.html'), 'w', encoding='utf-8')
        f.write(config.start_html(title, js))
        if readme:
            f.write('<div class="readme">%s</div>\n' % readme)
        if files:
            f.write('<ul>\n')
            for fn, path in files:
                if os.path.isdir(path):
                    shutil.copytree(path, os.path.join(outdir, fn))
                else:
                    shutil.copy(path, outdir)

                if not fn.startswith(','):
                    f.write('<li><a href="%s">%s</a></li>\n' % (fn, fn))
            f.write('</ul>\n')
        f.write('''
    <form id="form" action="%(cgi)s" method="post">
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
''' % {'cgi': config.get('puzzler', 'cgi_url'),
       'cat': cat,
       'points': points})
        f.write(config.end_html())

f = open(opts.keyfile, 'w', encoding='utf-8')
for key in keys:
    f.write('%s\t%s\t%s\n' % key)

