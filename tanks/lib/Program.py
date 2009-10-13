"""<H2>Introduction</H2>
You are the proud new operator of a M-375 Pflanzarr Tank.  Your tank is 
equipped with a powerful laser cannon, independently rotating turret
section, up to 10 enemy detection sensors, and a standard issue NATO hull.
Unfortunately, it lacks seats, and thus must rely own its own wits and your
skills at designing those wits to survive.  

<H2>Programming Your Tank</H2>
Your tanks are programmed using the Super Useful Command and Kontrol language,
the very best in laser tank AI languages.  It includes amazing features such 
as comments (Started by a #, ended at EOL), logic, versatility, and 
semi-colons (all lines must end in one).  As with all new military systems 
it utilizes only integers; we must never rest in our
diligence against the communist floating point conspiracy.  Whitespace is 
provided by trusted contractors, and should never interfere with operations.
<P>
Your program should be separated into Setup and AI commands.  The definitions 
section lets you designated the behaviors of its sensors and memory.  
Each setup command must begin with a '>'. Placing setup commands after 
the first AI command is a violation of protocol.  
Here are some examples of correct setup commands:
<pre class="docs">>addsensor(80, 90, 33);
>addsensor(50, 0, 10, 1);
>addtimer(3);</pre>

The AI section will act as the brain of your tank.  Each AI line is 
separated into a group of conditions functions and a group of action 
functions.  If all the conditions are satisfactory (true), all of the actions
are given as orders.  Conditions are separated by ampersands, actions separated 
by periods. Here are some examples of AI commands:
<pre class="docs">
sensor(1) & sensor(2) & fireready() : fire();
sensor(0,0)&sin(5): move(40, 30) . turretcw(50);
sensor(4) & random(4,5) : led(1).settoggle(0,1);</pre>

Your tank will check its program each turn, and attempt to the best of its
abilities to carry out its orders (or die trying).  Like any military mind, 
your tank may receive a plethora of often conflicting orders and information.  
This a SMART TANK, however.  It knows that the proper thing to do with each
subsystem is to have that subsystem follow only the last order given each turn.
"""

import conditions
import actions
import setup

class Statement(object):
    """Represents a single program statement.  If all the condition Functions
    evaluate to True, the actions are all executed in order."""

    def __init__(self, lineNum, line, conditions, actions):
        self.lineNum = lineNum
        self.line = line
        self._conditions = conditions
        self._actions = actions

    def __call__(self, tank):
        success = True
        for condition in self._conditions:
            if not condition(tank):
                success = False
                break

        if success:    
            for action in self._actions:
                action(tank)

class Program(object):
    """This parses and represents a Tank program."""
    CONDITION_SEP = '&'
    ACTION_SEP = '.'

    def __init__(self, text):
        """Initialize this program, parsing the given text."""
        self.errors = []

        self._program, self._setup = self._parse(text)
    
    def setup(self, tank):
        """Execute all the setup actions."""
        for action in self._setup:
            try:
                action(tank)
            except Exception, msg:
                self.errors.append("Bad setup action, line %d, msg: %s" % \
                                   (action.lineNum, msg))

    def __call__(self, tank):
        """Execute this program on the given tank."""
        for statement in self._program:
            try:
                statement(tank)
            except Exception, msg:
                self.errors.append('Error executing program. \n'
                                   '(%d) - %s\n'
                                   'msg: %s\n' % 
                                   (statement.lineNum, statement.line, msg) )

    def _parse(self, text):
        """Parse the text of the given program."""
        program = []
        setup = []
        inSetup = True
        lines = text.split(';')
        lineNum = 0
        for line in lines:
            lineNum = lineNum + 1

            originalLine = line

            # Remove Comments
            parts = line.split('\n')
            for i in range(len(parts)):
                comment = parts[i].find('#')
                if comment != -1:
                    parts[i] = parts[i][:comment]
            # Remove all whitespace
            line = ''.join(parts)
            line = line.replace('\r', '')
            line = line.replace('\t', '')
            line = line.replace(' ', '')
            
            if line == '':
                continue

            if line.startswith('>'):
                if inSetup:
                    if '>' in line[1:] or ':' in line:
                        self.errors.append('(%d) Missing semicolon: %s' % 
                                           (lineNum, line))
                        continue

                    try:
                        setupAction = self._parseSection(line[1:], 'setup')[0]
                        setupAction.lineNum = lineNum
                        setup.append(setupAction)
                    except Exception, msg:
                        self.errors.append('(%d) Error parsing setup line: %s'
                                           '\nThe error was: %s' %
                                           (lineNum, originalLine, msg))

                    continue
                else: 
                    self.errors.append('(%d) Setup lines aren\'t allowed '
                                       'after the first command: %s' %
                                       (lineNum, originalLine))
            else:
                # We've hit the first non-blank, non-comment, non-setup
                # line
                inSetup = False

            semicolons = line.count(':')
            if semicolons > 1:
                self.errors.append('(%d) Missing semicolon: %s' %
                                   (lineNum, line))
                continue
            elif semicolons == 1:
                conditions, actions = line.split(':')
            else:
                self.errors.append('(%d) Invalid Line, no ":" seperator: %s'%
                                     (lineNum, line) )

            try:
                conditions = self._parseSection(conditions, 'condition')
            except Exception, msg:
                self.errors.append('(%d) %s - "%s"' % 
                                    (lineNum, msg, line) )
                continue

            try:
                actions = self._parseSection(actions, 'action')
            except Exception, msg:
                self.errors.append('(%d) %s - "%s"' % 
                                    (lineNum, msg, originalLine) )
                continue
            program.append(Statement(lineNum, line, conditions, actions))

        return program, setup

    def _parseSection(self, section, sectionType):
        """Parses either the action or condition section of each command.
@param section: The text of the section of the command to be parsed.
@param sectionType: The type of section to be parsed.  Should be:
                    'condition', 'action', or 'setup'.
@raises ValueError: Raises ValueErrors when parsing errors occur.
@returns: Returns a list of parsed section components (Function objects).
        """

        if sectionType == 'condition':
            parts = section.split(self.CONDITION_SEP)
            functions = conditions.conditions
            if section == '':
                return []
        elif sectionType == 'action':
            parts = section.split(self.ACTION_SEP)
            functions = actions.actions
            if section == '':
                raise ValueError("The action section cannot be empty.")
        elif sectionType == 'setup':
            parts = [section]
            functions = setup.setup
        else:
            raise ValueError('Invalid section Type - Contact Contest Admin')

        parsed = [] 
        for part in parts:

            pos = part.find('(')
            if pos == -1:
                raise ValueError("Missing open paren in %s: %s" % 
                                 (sectionType, part) )
            funcName = part[:pos]

            if funcName not in functions:
                raise ValueError("%s function %s is not accepted." %
                                 (sectionType.capitalize(), funcName) )

            if part[-1] != ')':
                raise ValueError("Missing closing paren in %s: %s" %
                                 (condition, sectionType) )

            args = part[pos+1:-1]
            if args != '':
                args = args.split(',')
                for i in range(len(args)):
                    args[i] = int(args[i])
            else:
                args = []

            parsed.append(functions[funcName](*args))

        return parsed
