#! /usr/bin/env python3

##
## XXX: Move more of this into game
##

import game

class Roshambo(game.Game):
    def setup(self):
        print("Hello from setup")
        self.other_move = None

    def make_move(self, player, move):
        print(self.other_move, player, move)
        if self.other_move:
            other_player, other_move = self.other_move
            moves = (move, other_move)
            if move in (('rock', 'scissors'),
                        ('scissors', 'paper'),
                        ('paper', 'rock')):
                # Player wins
                self.declare_winner(player)
            else:
                self.declare_winner(other_player)
            other_player.unblock()
        else:
            self.other_move = (player, move)
            player.block()

    def handle(self, player, cmd, args):
        if cmd in ('rock', 'scissors', 'paper'):
            self.make_move(player, cmd)
        else:
            raise ValueError('Invalid command')

def main():
    game.run(2, Roshambo, 5388, b'roshambo:::984233f357ecac03b3e38b9414cd262b')

if __name__ == '__main__':
    main()

