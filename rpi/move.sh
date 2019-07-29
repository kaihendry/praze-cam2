#!/bin/bash

sdir=${sdir:-/home/pi/amcrest/AMC02785_1ACA93} # source directory
wdir=${wdir:-/home/pi/cam} # web directory

test -d $sdir || exit

np="$(mktemp -d)/copy-$$"
mkfifo "$np" || exit

inotifywait -r -m -e MOVED_TO "$sdir" > "$np" & ipid=$!

trap "kill $ipid; rm -f $np" EXIT

while read -r dir _ fn
do

	od=$wdir/$(date +%Y-%m-%d) # output directory with date prefixed by user's home dir
	mkdir -p $od || true

	case "${fn##*.}" in

	mp4)
		if FFREPORT=file=/tmp/htmlvideo.log:level=32 ffmpeg -v panic -i $dir/$fn -c:v copy -tag:v hvc1 $od/$fn < /dev/null
		then
			rm -vf $dir/$fn
			poster="$od/$(basename $fn mp4)jpg"
			ffmpeg -y -i $od/$fn -f mjpeg -vframes 1 -ss 10 "$poster"
			jpegtmp=$(mktemp --suffix=.jpg)
			/opt/mozjpeg/bin/cjpeg -quality 80 "$poster" > $jpegtmp
			echo "Squashing jpeg $(du -h "$poster" "$jpegtmp")"
			mv $jpegtmp "$poster"
		fi
		;;

	jpg)
		ofn=$od/$fn # output file name
		echo $fn is an image, moving to $ofn
		mv $dir/$fn $ofn && chmod a+r $ofn
		;;

	*)
		echo Unknown type: ${dir}${fn}
		;;

	esac

done < "$np"
