#! /usr/bin/python3

## Course assignments

import csv
import smtplib

msg = '''From: Neale Pickett <neale@lanl.gov>
To: %(recip)s
Subject: Tracer FIRE 4 course assignment: %(course)s

Hello!  Your course assignment for Tracer FIRE 4 is:

    %(course)s

Please see http://csr.lanl.gov/tf/tf4.html for information on
what you need to bring to the course.

Course questions should be directed to the appropriate instructor:

    Network RE:             Neale Pickett <neale@lanl.gov>
    Malware RE:             Danny Quist <dquist@lanl.gov>
    Host Forensics:         Kevin Nauer <ksnauer@sandia.gov>
    Incident Coordination:  Alex Kent <alex@lanl.gov>

General questions about Tracer FIRE may be sent to
Neale Pickett <neale@lanl.gov>

Remember: the exercise network should be considered
hostile!  Do not bring anything sensitive on your laptop,
and make sure you back everything up.

Looking forward to seeing you in Santa Fe next week,

-- 
Neale Pickett <neale@lanl.gov>
Advanced Computing Solutions, Los Alamos National Laboratory
'''

limits = {'Malware RE': 26,
          'Network RE': 40}
assignments = {}

assigned = set(l.strip() for l in open('assigned.txt'))

c = csv.reader(open('/tmp/g.csv'))
c.__next__()
for row in c:
    assert '@' in row[2]
    t = row[5]
    if (len(assignments.get(t, '')) == limits.get(t, 50)):
        if (row[6] == row[5]):
            print("Jackass detected: %s" % row[2])
        t = row[6]
    l = assignments.setdefault(t, [])
    l.append(row)

s = smtplib.SMTP('mail.lanl.gov')
for t in ('Incident Coordinator', 'Network RE', 'Malware RE', 'Forensics'):
    print('%s (%s)' % (t, len(assignments[t])))
    for row in assignments[t]:
        e = row[2]
        if e in assigned:
            print('    %s' % e)
        else:
            print(' *  %s' % e)
            ret = s.sendmail('neale@lanl.gov', [e], msg % {'course': t, 'recip': e})
            if ret:
                print(' ==>    %s' % ret)
            else:
                assigned.add(e)
s.quit()

a = open('assigned.txt', 'w')
for e in assigned:
    a.write('%s\n' % e)
