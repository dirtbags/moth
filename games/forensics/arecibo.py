#! /usr/bin/env python3

#       0        1         2         3         4
#       1234567890123456789012345678901234567890123
#       ||| | |   | |   | |   |     | |     |   | |
msg = (' ####  ####   #   ###   ###  ###    # ###  '
       ' #   # #   # ##  #     #        #  #     # '
       ' #   # ####   #  ####  ####    #  #     #  '
       ' #   # #   #  #  #   # #   #    #  #     # '
       ' ####  ####  ###  ###   ###  ###    # ###  ')

msg = msg.replace('#', '0')
msg = msg.replace(' ', '1')
num = int(msg, 2)
print(num)
