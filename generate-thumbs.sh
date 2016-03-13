#!/bin/bash

SOURCE=~/Dropbox/photos-site

for folder in $SOURCE/*; do
    foldername=${folder##*/}
    mkdir -p photos/$foldername

    for file in $folder/*; do
        filename=${file##*/}
        out="photos/$foldername/$filename";
        test -e $out || convert -thumbnail 400^ -gravity center -crop 400x200+0+0 -strip -quality 86 $file $out;
    done
done

