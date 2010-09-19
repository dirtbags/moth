#! /bin/sh -e

die () {
    echo "ERROR ERROR: $*"
    exit 1
}

CTF_BASE=/tmp/ctf-test.$$ export CTF_BASE
trap "rm -rf $CTF_BASE" 0
mkdir $CTF_BASE

# Some skeletal structure
mkdir -p $CTF_BASE/points.new

# Set up some packages
for cat in cat1 cat2 cat3; do
    mkdir -p $CTF_BASE/packages/$cat
    cat >$CTF_BASE/packages/$cat/answers.txt <<EOF
10 ${cat}answer10
20 ${cat}answer20
30 ${cat}answer30
EOF
    for i in 10 20 30; do
        mkdir -p $CTF_BASE/packages/$cat/puzzles/$i
    done
done

# Set up some teams
mkdir -p $CTF_BASE/teams/names
mkdir -p $CTF_BASE/teams/colors
for team in team1 team2 team3; do
    hash=$(bin/register $team | awk '{print $NF;}')
done


##
## Puzzler tests
##

if src/puzzles.cgi | grep 20; then
    die "20 points puzzles shouldn't show up here"
fi

if src/puzzler.cgi t=$hash c=cat1 p=10 a=cat1answer20 | grep -q 'awarded'; then
    die "Awarded points with wrong answer"
fi

if src/puzzler.cgi t=$hash c=cat2 p=10 a=cat1answer10 | grep -q 'awarded'; then
    die "Awarded points with wrong category"
fi

if src/puzzler.cgi t=$hash c=cat1 p=20 a=cat1answer10 | grep -q 'awarded'; then
    die "Awarded points with wrong point value"
fi

if src/puzzler.cgi t=merfmerfmerfmerf c=cat2 p=10 a=cat1answer10 | grep -q 'awarded'; then
    die "Awarded points with bad team"
fi

if ! src/puzzler.cgi t=$hash c=cat1 p=10 a=cat1answer10 | grep -q 'awarded 10'; then
    die "Didn't award points for correct answer"
fi

if ! src/puzzles.cgi | grep -q 20; then
    die "20 point answer didn't show up"
fi

if src/puzzler.cgi t=$hash c=cat1 p=10 a=cat1answer10 | grep -q 'awarded 10'; then
    die "Awarded same points twice"
fi

##
## Scoreboard tests
##

if ! cat $CTF_BASE/points.new/* | bin/scoreboard | grep -q 'total.*team3: 1'; then
    die "Scoreboard total incorrect"
fi

if ! cat $CTF_BASE/points.new/* | bin/scoreboard | grep -q 'cat1.*team3: 10'; then
    die "Scoreboard cat1 points incorrect"
fi

##
## Token tests
##

mkdir -p $CTF_BASE/token.keys
echo -n '0123456789abcdef' > $CTF_BASE/token.keys/tokencat

mkfifo $CTF_BASE/nancy
src/tokencli tokencat $CTF_BASE/token.keys/tokencat < $CTF_BASE/nancy 3>$CTF_BASE/t | src/in.tokend > $CTF_BASE/nancy

if ! grep -q 'tokencat:x....-....x' $CTF_BASE/tokens.db; then
    die "in.tokend didn't write to database"
fi

if src/claim.cgi t=lalalala k=$(cat $CTF_BASE/tokens.db) | grep -q success; then
    die "claim.cgi gave points to a bogus team"
fi

if src/claim.cgi t=$hash k=tokencat:xanax-xanax | grep -q success; then
    die "claim.cgi gave points for a bogus token"
fi

if ! src/claim.cgi t=$hash k=$(cat $CTF_BASE/t) | grep -q success; then
    die "claim.cgi didn't give me any points"
fi

if src/claim.cgi t=$hash k=$(cat $CTF_BASE/t) | grep -q success; then
    die "claim.cgi gave me points twice for the same token"
fi

if ! [ -f $CTF_BASE/points.new/*.$hash.tokencat.1 ]; then
    die "claim.cgi didn't actually record any points"
fi

echo "All tests passed!  You're the best programmer ever!"
