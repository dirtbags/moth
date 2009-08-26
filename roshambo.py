#! /usr/bin/env python3

import game

class Roshambo(game.TurnBasedGame):
    def setup(self):
        self.moves = []

    def calculate_moves(self):
        moves = [m[1] for m in self.moves]
        if moves[0] == moves[1]:
            self.moves[0][0].write('tie')
            self.moves[1][0].write('tie')
            self.moves = []
        elif moves in (('rock', 'scissors'),
                       ('scissors', 'paper'),
                       ('paper', 'rock')):
            # First player wins
            self.declare_winner(self.moves[0][0])
        else:
            self.declare_winner(self.moves[1][0])

    def make_move(self, player, move):
        self.moves.append((player, move))
        self.end_turn(player)

    def do_rock(self, player, args):
        self.make_move(player, 'rock')

    def do_scissors(self, player, args):
        self.make_move(player, 'scissors')

    def do_paper(self, player, args):
        self.make_move(player, 'paper')


def main():
    game.run(2, Roshambo, 5388, b'roshambo:::984233f357ecac03b3e38b9414cd262b')

if __name__ == '__main__':
    main()

