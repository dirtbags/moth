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
                   'css_url': 'ctf.css',
                   'diasbled_dir': 'disabled',
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
    config = {'global':
                  {'data_dir': '/var/lib/ctf',
                   'base_url': '/',
                   'css_url': '/ctf.css',
                   'disabled_dir': '/var/lib/ctf/disabled',
                   'flags_dir': '/var/lib/ctf/flags',
                   'house_team': 'dirtbags',
                   'passwd': '/var/lib/ctf/passwd',
                   'team_colors': team_colors,
                   },
              'puzzler':
                  {'dir': '/usr/lib/www/puzzler',
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
