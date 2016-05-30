#!/bin/bash

SOURCE=~/Dropbox/photos-site
TEMPLATE='---
album_title: XXX
album_date: YYY
date: YYYY-MM-DD
title: Photos of XXX — YYY
url: /photos/$foldername/
photos:
$PHOTOS
---'

PHOTO_PRINTF='- file: %f\n  alt: Unknown\n  caption: Unknown\n'

for folder in $SOURCE/*; do
    foldername=${folder##*/}
    mkdir -p photos/$foldername

    for file in $folder/*; do
        filename=${file##*/}
        out="site/static/photos/$foldername/$filename";
        test -e "$out" || convert -thumbnail 400^ -gravity center -crop 400x200+0+0 -strip -quality 86 "$file" "$out";
    done

    if [ ! -e "site/content/photos/$foldername.html" ]; then
        export PHOTOS=$(find "site/static/photos/$foldername" -type f -printf "$PHOTO_PRINTF")
        envsubst <<< "$TEMPLATE" > "site/content/photos/$foldername.html"
    fi
done
