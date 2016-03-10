#!/bin/bash

for file in `find photos -type f | grep -v .thumb. | grep -v .html`; do
    out="${file%.*}.thumb.${file##*.}";
    test -e $out || convert -thumbnail 400 $file $out;
done

