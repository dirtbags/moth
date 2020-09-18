#! /bin/sh -e

fail () {
    echo "$@" 1>&2
    exit 1
}

case "$ACTION:$CAT:$POINTS" in
    inventory::)
        cat <<EOT
{
    "pategory": [1, 2, 3, 4, 5, 10, 20, 300],
    "nealegory": [1, 3, 2]
}
EOT
        ;;
    open:*:*)
        case "$CAT:$POINTS:$FILENAME" in
            *:*:moo.txt)
                echo "Moo."
                ;;
            *)
                fail "Cannot open: $FILENAME"
                ;;
        esac
        ;;
    answer:pategory:1)
        if [ "$ANSWER" = "answer" ]; then
            echo "correct"
        else
            echo "Sorry, wrong answer."
        fi
        ;;
    answer:pategory:2)
        fail "Internal error"
        ;;
    *)
        fail "ERROR: Unknown action: $action"
        ;;
esac
