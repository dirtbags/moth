POINT_VALUES = [1,2,3,4,5]

def pointvals():
    return POINT_VALUES

def make(points, puzzle):
    if points not in POINT_VALUES:
        return None

    puzzle.answers.append(points)
    return puzzle
