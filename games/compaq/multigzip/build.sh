#! /bin/sh

key='PC LOAD LETTER'
lorem='Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc fermentum magna ut posuere.'

echo -n "$key" | gzip | split -b 1
rm -f multi.gz
for i in x??; do 
    nfn=$(xxd -c 256 -p $i)
    echo -n "$lorem" > $nfn
    gzip -c -N $nfn >> multi.gz
    rm $nfn
    rm $i
done
