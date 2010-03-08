#! /usr/bin/python

"""A shitty FORTH interpreter

15:58 <SpaceHobo> WELCOME TO FORF!
15:58 <SpaceHobo> *PUNCH*
"""

import operator

class ParseError(Exception):
    pass

class Overflow(Exception):
    pass

class Underflow(Exception):
    pass

class Stack:
    def __init__(self, init=None, size=50):
        self.size = size
        self.stack = init or []

    def __str__(self):
        if not self.stack:
            return '{}'
        guts = ' '.join(repr(i) for i in self.stack)
        return '{ %s }' % guts
    __repr__ = __str__

    def push(self, *values):
        for val in values:
            if len(self.stack) == self.size:
                raise Overflow()
            self.stack.append(val)

    def extend(self, other):
        self.stack.extend(other.stack)

    def dup(self):
        return Stack(init=self.stack[:], size=self.size)

    def pop(self):
        if not self.stack:
            raise Underflow()
        return self.stack.pop()

    def mpop(self, n):
        return [self.pop() for i in range(n)]

    def __nonzero__(self):
        return bool(self.stack)


class Environment:
    def __init__(self, ticks=2000, codelen=500):
        self.ticks = ticks
        self.codelen = codelen
        self.registers = [0] * 10
        self.unfuncs = {'~' : operator.inv,
                        '!' : operator.not_,
                        'abs': operator.abs,
                        }
        self.binfuncs = {'+' : operator.add,
                         '-' : operator.sub,
                         '*' : operator.mul,
                         '/' : operator.div,
                         '%' : operator.mod,
                         '**': operator.pow,
                         '&' : operator.and_,
                         '|' : operator.or_,
                         '^' : operator.xor,
                         '<<': operator.lshift,
                         '>>': operator.rshift,
                         '>' : operator.gt,
                         '>=': operator.ge,
                         '<' : operator.lt,
                         '<=': operator.le,
                         '=' : operator.eq,
                         '<>': operator.ne,
                         '!=': operator.ne,
                         }
        self.data = Stack()

    def get(self, s):
        unfunc = self.unfuncs.get(s)
        if unfunc:
            return self.apply_unfunc(unfunc)

        binfunc = self.binfuncs.get(s)
        if binfunc:
            return self.apply_binfunc(binfunc)

        try:
            return getattr(self, 'cmd_' + s)
        except AttributeError:
            return None

    def apply_unfunc(self, func):
        """Apply a unary function"""

        def f(data):
            a = data.pop()
            data.push(int(func(a)))
        return f

    def apply_binfunc(self, func):
        """Apply a binary function"""

        def f(data):
            a = data.pop()
            b = data.pop()
            data.push(int(func(b, a)))
        return f

    def run(self, s):
        self.parse_str(s)
        self.eval()

    def parse_str(self, s):
        tokens = s.strip().split()
        tokens.reverse()        # so .parse can tokens.pop()
        self.code = self.parse(tokens)

    def parse(self, tokens, token=0, depth=0):
        if depth > 4:
            raise ParseError('Maximum recursion depth exceeded at token %d' % token)
        code = []
        while tokens:
            val = tokens.pop()
            token += 1
            f = self.get(val)
            if f:
                code.append(f)
            elif val == '(':
                # Comment
                while val != ')':
                    val = tokens.pop()
                    token += 1
            elif val == '{}':
                # Empty block
                code.append(Stack())
            elif val == '{':
                block = self.parse(tokens, token, depth+1)
                code.append(block)
            elif val == '}':
                break
            else:
                # The only literals we support are ints
                try:
                    code.append(int(val))
                except ValueError:
                    raise ParseError('Invalid literal at token %d (%s)' % (token, val))
            if len(code) > self.codelen:
                raise ParseError('Code stack overflow')
        # Reverse so we can .pop()
        code.reverse()
        return Stack(code, size=self.codelen)

    def eval(self):
        ticks = self.ticks
        code_orig = self.code.dup()
        while self.code and ticks:
            ticks -= 1
            val = self.code.pop()
            try:
                if callable(val):
                    val(self.data)
                else:
                    self.data.push(val)
            except Underflow:
                self.err('Stack underflow at proc %r' % (val))
            except Overflow:
                self.err('Stack overflow at proc %r' % (val))
        if self.code:
            self.err('Ran out of ticks!')
        self.code = code_orig

    def err(self, msg):
        print 'Error: %s' % msg

    def msg(self, msg):
        print msg

    ##
    ## Commands
    ##
    def cmd_print(self, data):
        a = data.pop()
        self.msg(a)

    def cmd_dumpstack(self, data):
        a = data.pop()
        self.msg('(dumpstack %d) %r' % (a, data.stack))

    def cmd_dumpmem(self, data):
        a = data.pop()
        self.msg('(dumpmem %d) %r' % (a, self.registers))

    def cmd_exch(self, data):
        a, b = data.mpop(2)
        data.push(a, b)

    def cmd_dup(self, data):
        a = data.pop()
        data.push(a, a)

    def cmd_pop(self, data):
        data.pop()

    def cmd_store(self, data):
        a, b = data.mpop(2)
        self.registers[a % 10] = b

    def cmd_fetch(self, data):
        a = data.pop()
        data.push(self.registers[a % 10])

    ##
    ## Evaluation commands
    ##
    def eval_block(self, block):
        try:
            self.code.extend(block)
        except TypeError:
            # If it's not a block, just append it
            self.code.push(block)

    def cmd_if(self, data):
        block = data.pop()
        cond = data.pop()
        if cond:
            self.eval_block(block)

    def cmd_ifelse(self, data):
        elseblock = data.pop()
        ifblock = data.pop()
        cond = data.pop()
        if cond:
            self.eval_block(ifblock)
        else:
            self.eval_block(elseblock)

    def cmd_eval(self, data):
        # Interestingly, this is the same as "1 exch if"
        block = data.pop()
        self.eval_block(block)

    def cmd_call(self, data):
        # Shortcut for "fetch eval"
        self.cmd_fetch(data)
        self.cmd_eval(data)


def repl():
    env = Environment()
    while True:
        try:
            s = raw_input('>8[= =] ')
        except (KeyboardInterrupt, EOFError):
            print
            break
        try:
            env.run(s)
            print env.data
        except ParseError, err:
            print r'    \ nom nom nom, %s!' % err
    print r'    \ bye bye!'

if __name__ == '__main__':
    import sys
    import time
    try:
        import readline
    except ImportError:
        pass

    if len(sys.argv) > 1:
        s = open(sys.argv[1]).read()
        env = Environment()
        begin = time.time()
        env.run(s)
        end = time.time()
        elapsed = end - begin
        print 'Evaluated in %.2f seconds' % elapsed
    else:
        print 'WELCOME TO FORF!'
        print '*PUNCH*'
        repl()
