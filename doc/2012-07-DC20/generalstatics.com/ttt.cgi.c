#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "cgi.h"

const char *b64_aleph = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@";

typedef int board_t[3][3];

void
b64_of_board(char *out, board_t board)
{
    int y, x;

    for (y = 0; y < 3; y += 1) {
        int acc = 0;

        for (x = 0; x < 3; x += 1) {
            acc <<= 2;
            acc += board[y][x];
        }
        out[y] = b64_aleph[acc];
    }
}

void
board_of_b64(board_t out, char *b64)
{
    int y, x;

    for (y = 0; y < 3; y += 1) {
        char *p = strchr(b64_aleph, b64[y]);
        int acc = 0;

        if (p) {
            acc = p - b64_aleph;
        }

        for (x = 2; x >= 0; x -= 1) {
            out[y][x] = acc & 3;
            acc >>= 2;
        }
    }
}

void
print_board(board_t board)
{
    int y, x;

    for (y = 0; y < 3; y += 1) {
        for (x = 0; x < 3; x += 1) {
            printf("%d", board[y][x]);
        }
        printf("\n");
    }
}

int
winner(board_t board)
{
    int i, j, k;

    for (i = 0; i < 3; i += 1) {
        for (k = 0; k < 3; k += 1) {
            int winner = -1;

            for (j = 0; j < 3; j += 1) {
                int b;

                switch (k) {
                    case 0:
                        b = board[i][j];
                        break;
                    case 1:
                        b = board[j][i];
                        break;
                    case 2:
                        /* This will happen 3× as often as it needs to.  Who cares. */
                        b = board[j][j];
                        break;
                }

                if (winner == -1) {
                    winner = b;
                } else if (winner != b) {
                    winner = -1;
                    break;
                }
            }
            if (winner > 0) {
                return winner;
            }
        }
    }

    return 0;
}

void
claim(board_t board, int x, int y, int whom)
{
    int prev = board[x][y];
    int i;

    if (prev == whom) {
        return;
    }

    for (i = 0; i < 9; i += 1) {
        if (! board[i/3][i%3]) {
            board[i/3][i%3] = prev;
            break;
        }
    }

    board[x][y] = whom;
}

void
make_move(board_t board)
{
    switch (winner(board)) {
        case 1:
            printf("A WINNER IS YOU\n");
            exit(0);
        case 2:
            /* I win; we can keep playing though, because I (neale)
               don't want to write any more code to handle this. */
            break;
        case 3:
            printf("A WINNER IS WHO?\n");
            exit(1);
    }

    /* Reserve our final space */
    claim(board, 2, 2, 0);

    /* First move */
    if (board[1][1] != 2) {
        claim(board, 1, 1, 2);
        return;
    }

    /* Second move */
    if (board[0][0] != 2) {

        /* Prevent them from winning legally */
        if (board[0][2]) {
            claim(board, 1, 2, 0);
        }
        if (board[2][0]) {
            claim(board, 2, 1, 0);
        }
        claim(board, 0, 0, 2);
        return;
    }

    /* Third move */
    claim(board, 2, 2, 2);
}

        
int
main(int argc, char *argv[])
{
    char b64[4] = {0};
    board_t board = {0};

    if (-1 == cgi_init(argv)) {
        return 0;
    }

    while (1) {
        size_t len;
        char   key[20];

        len = cgi_item(key, sizeof key);
        if (0 == len) break;
        switch (key[0]) {
            case 'b':
                cgi_item(b64, sizeof b64);
                break;
            default:
                cgi_item(key, 0);
                break;
        }
    }

    printf("Content-type: text/plain\r\n\r\n");
    board_of_b64(board, b64);
    make_move(board);
    b64_of_board(b64, board);
    fwrite(b64, 1, 3, stdout);

    return 0;
}

