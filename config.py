#! /usr/bin/env python3

import os

if 'home' in os.environ.get('SCRIPT_FILENAME', ''):
    # We're a CGI running out of someone's home directory
    config = {'global':
                  {'data_dir': '.',
                   'base_url': '.',
                   'css_url': 'ctf.css',
                   'diasbled_dir': 'disabled'
                   },
              'puzzler':
                  {'dir': 'puzzles',
                   'ignore_dir': 'puzzler.ignore',
                   'cgi_url': 'puzzler.cgi',
                   'base_url': 'puzzler',
                   'keys_file': 'puzzler.keys',
                   },
              }
else:
    # An actual installation
    config = {'global':
                  {'data_dir': '/var/lib/ctf',
                   'base_url': '/',
                   'css_url': '/ctf.css',
                   'disabled_dir': '/var/lib/ctf/disabled',
                   },
              'puzzler':
                  {'dir': '/usr/lib/www/puzzler',
                   'ignore_dir': '/var/lib/ctf/puzzler.ignore',
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
