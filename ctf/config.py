#! /usr/bin/env python3

import os

team_colors = ['F0888A', '88BDF0', '00782B', '999900', 'EF9C00',
               'F4B5B7', 'E2EFFB', '89CA9D', 'FAF519', 'FFE7BB',
               'BA88F0', '8DCFF4', 'BEDFC4', 'FFFAB2', 'D7D7D7',
               'C5B9D7', '006189', '8DCB41', 'FFCC00', '898989']

if 'home' in os.environ.get('SCRIPT_FILENAME', ''):
    # We're a CGI running out of someone's home directory
    config = {'global':
                  {'data_dir': '.',
                   'base_url': '.',
                   'css_url': '/~neale/ctf/ctf.css',
                   'disabled_dir': 'disabled',
                   'flags_dir': 'flags',
                   'house_team': 'dirtbags',
                   'passwd': 'passwd',
                   'team_colors': team_colors,
                   },
              'puzzler':
                  {'dir': 'puzzles',
                   'cgi_url': 'puzzler.cgi',
                   'base_url': 'puzzler',
                   'keys_file': 'puzzler.keys',
                   },
              }
else:
    # An actual installation
    config = {
        'global':
            {
            'data_dir': '/var/lib/ctf',
            'base_url': '/',
            'css_url': '/ctf.css',
            'disabled_dir': '/var/lib/ctf/disabled',
            'flags_dir': '/var/lib/ctf/flags',
            'house_team': 'dirtbags',
            'passwd': '/var/lib/ctf/passwd',
            'team_colors': team_colors,
            },
        'pollster':
            {
            'poll_interval': 60,
            'poll_timeout': 0.5,
            'heartbeat_dir': '/var/lib/pollster',
            'results': '/var/lib/pollster/status.html',
            },
        'puzzler':
            {
            'dir': '/usr/lib/www/puzzler',
            'cgi_url': '/puzzler.cgi',
            'base_url': '/puzzler',
            'keys_file': '/usr/lib/ctf/puzzler.keys',
            },
        }

def get(section, key):
    return config[section][key]

disabled_dir = get('global', 'disabled_dir')
data_dir = get('global', 'data_dir')
base_url = get('global', 'base_url')
css = get('global', 'css_url')

def disabled(cat):
    path = os.path.join(disabled_dir, cat)
    return os.path.exists(path)

def enabled(cat):
    return not disabled(cat)

def datafile(filename):
    return os.path.join(data_dir, filename)

def url(path):
    return base_url + path

def start_html(title, hdr='', cls='', links=[], links_title=None):
    ret = []
    if os.environ.get('GATEWAY_INTERFACE'):
        ret.append('Content-type: text/html')
        ret.append('')
    ret.append('''<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC
  "-//W3C//DTD XHTML 1.0 Strict//EN"
  "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <title>%(title)s</title>
    <link rel="stylesheet" href="%(css)s" type="text/css" />
    %(hdr)s
  </head>
  <body class="%(class)s">
    <h1>%(title)s</h1>
    <div id="navigation">
      <ul>
        <li><a href="%(base)s">Home</a></li>
        <li><a href="%(base)sintro.html">Intro/Rules</a></li>
        <li><a href="%(base)sservices.html">Svc flags</a></li>
        <li><a href="%(base)s/tanks/results.cgi">Tanks</a></li>
        <li><a href="%(base)spuzzler.cgi">Puzzles</a></li>
        <li><a href="%(base)sscoreboard.cgi">Scoreboard</a></li>
      </ul>
''' % {'title': title,
       'css': css,
       'hdr': hdr,
       'base': base_url,
       'class': cls})
    if links:
        if links_title:
            ret.append('<h3>%s</h3>' % links_title)
        else:
            ret.append('<hr/>')
        for url, name in links:
            ret.append('<li><a href="%s">%s</a></li>' % (url, name))
    ret.append('    </div>')
    return '\n'.join(ret)

def end_html():
    return '</body></html>'
