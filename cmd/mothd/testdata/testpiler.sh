#! /bin/sh -e

case "$action:$cat:$points" in
    inventory::)
        echo "pategory 1 2 3 4 5 10 20 300"
        echo "nealegory 1 2 3"
        ;;
    open:*:*)
        if [ "$path" = "moo.txt" ]; then
            echo "Moo."
        else
            cat $cat_$points_$path
        fi
        ;;
    answer:pategory:1)
        if [ "$answer" = "answer" ]; then
            echo "correct"
        else
            echo "Sorry, wrong answer."
        fi
        ;;
    answer:pategory:2)
        echo "Internal error" 1>&2
        exit 1
        ;;
    *)
        echo "ERROR: Unknown action: $action" 1>&2
        exit 1
        ;;
esac
