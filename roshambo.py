#! /usr/bin/env python3

import game

class Roshambo(game.TurnBasedGame):
    def setup(self):
        self.moves = []

    def calculate_moves(self):
        players = [m[0] for m in self.moves]
        moves = [m[1] for m in self.moves]
        if moves[0] == moves[1]:
            players[0].write('tie')
            players[1].write('tie')
            self.moves = []
        elif moves in (['rock', 'scissors'],
                       ['scissors', 'paper'],
                       ['paper', 'rock']):
            # First player wins
            self.declare_winner(players[0])
        else:
            self.declare_winner(players[1])

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
    game.run(Roshambo, 5388, b'roshambo:::984233f357ecac03b3e38b9414cd262b', 2)

if __name__ == '__main__':
    main()

