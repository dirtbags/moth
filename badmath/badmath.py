import random
import math

OPS = [lambda a, b: a + b,
       lambda a, b: a - b,
       lambda a, b: a * b,
       lambda a, b: a // b,
       lambda a, b: a % b,
       lambda a, b: a ^ b,
       lambda a, b: a | b,
       lambda a, b: a & b,
       lambda a, b: max(a,b),
       lambda a, b: min(a,b),
       lambda a, b: a+b//2,
       lambda a, b: ~b,
       lambda a, b: a + b + 3,
       lambda a, b: max(a,b)//2,
       lambda a, b: min(a,b)*3,
       lambda a, b: a % 2,
       lambda a, b: int(math.degrees(b + a)),
       lambda a, b: ~(a & b),
       lambda a, b: ~(a ^ b),
       lambda a, b: a + b - a%b,
       lambda a, b: math.factorial(a)//math.factorial(a-b) if a > b else 0,
       lambda a, b: (b%a) * (a%b),
       lambda a, b: math.factorial(a)%b,
       lambda a, b: int(math.sin(a)*b),
       lambda a, b: b + a%2,
       lambda a, b: a - 1 + b%3,
       lambda a, b: a & 0xaaaa,
       lambda a, b: 5 if a == b else 6,
       lambda a, b: b % 17,
       lambda a, b: int( cos( math.radians(b) ) * a )]

SYMBOLS = '.,<>?/!@#$%^&*()_+="~|;:'
MAX = 100

PLAYER_DIR = ''

def mkPuzzle(lvl):
    """Make a puzzle.  The puzzle is a simple integer math equation.  The trick
    is that the math operators don't do what you might expect, and what they do
    is randomized each time (from a set list of functions).  The equation is
    evaluated left to right, with no other order of operations.
   
    The level determins both the length of the puzzle, and what functions are
    enabled. The number of operators is half the level+2, and the number of
    functions enabled is equal to the level.
    
    returns the key, puzzle, and the set of numbers used.
    """

    ops = OPS[:lvl + 1]
    length = (lvl + 2)//2

    key = {}
    
    bannedNums = set()

    puzzle = []
    for i in range(length):
        num = random.randint(1,MAX)
        bannedNums.add(num)
        puzzle.append( num )
        symbol = random.choice(SYMBOLS)
        if symbol not in key:
            key[symbol] = random.randint(0, len(ops) - 1)
        puzzle.append( symbol )
        
    num = random.randint(1,MAX)
    bannedNums.add(num)
    puzzle.append( num )
    
    return key, puzzle, bannedNums

def parse(puzzle):
    """Parse a puzzle string.  If the string contains symbols not in 
    SYMBOLS, a ValueError is raised."""

    parts = [puzzle]
    for symbol in SYMBOLS:
        newParts = []
        for part in parts:
            if  symbol in part:
                terms = part.split(symbol)
                newParts.append( terms.pop(0))
                while terms:
                    newParts.append(symbol)
                    newParts.append( terms.pop(0) )
            else:
                newParts.append(part)
        parts = newParts

    finalParts = []
    for part in parts:
        part = part.strip()
        if part in SYMBOLS:
            finalParts.append( part )
        else:
            try:
                finalParts.append( int(part) )
            except:
                raise ValueError("Invalid symbol: %s" % part)

    return finalParts

def solve(key, puzzle):

    puzzle = list(puzzle)
    stack = puzzle.pop(0)

    while puzzle:
        symbol = puzzle.pop(0)
        nextVal = puzzle.pop(0)
        op = OPS[key[symbol]]
        stack = op(stack, nextVal)

    return stack
