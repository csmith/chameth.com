#!/bin/bash

for file in `find photos -type f | grep -v .thumb. | grep -v .html`; do
    out="${file%.*}.thumb.${file##*.}";
    test -e $out || convert -thumbnail 400^ -gravity center -crop 400x200+0+0 $file $out;
done

