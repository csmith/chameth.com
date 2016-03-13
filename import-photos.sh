#!/bin/bash

SOURCE=~/Dropbox/photos-site
TEMPLATE='---
layout: photos
album_title: XXX
album_date: YYY
title: Photos of XXX — YYY · Chameth.com
photos:
$PHOTOS
---'

PHOTO_PRINTF='- file: %f\n  alt: Unknown\n  caption: Unknown\n'

for folder in $SOURCE/*; do
    foldername=${folder##*/}
    mkdir -p photos/$foldername

    for file in $folder/*; do
        filename=${file##*/}
        out="photos/$foldername/$filename";
        test -e $out || convert -thumbnail 400^ -gravity center -crop 400x200+0+0 -strip -quality 86 $file $out;
    done

    if [ ! -e "photos/$foldername/index.html" ]; then
        export PHOTOS=$(find "photos/$foldername" -type f -printf "$PHOTO_PRINTF")
        envsubst <<< "$TEMPLATE" > "photos/$foldername/index.html"
    fi
done
