# Private Camera viewer (S3 index)

Goal is to allow my family to view CCTV footage without requiring a complex NVR
or Cloud subscription.

Lists our S3 bucket by YYYY-MM-DD/ prefix for videos and displays them.

Shows yesterday's videos since they are only uploaded at 2AM.

# Upload

Please view the `rpi/` folder.

[Amcrest
IP4M-1026E](https://www.amazon.co.uk/Amcrest-Waterproof-Recording-4-Megapixel-IP4M-1026EB/dp/B073V6XMJN)
is configured to upload videos on motion detection via local FTP to a Raspberry
PI device.  `move.sh` fixes their H265 recordings for playback on Apple devices
as well as creating a video poster/thumbnail.

# Authentication

Once accessed over IP whitelist, a cookie is set that allows access on other IPs.

Cookie is secured by `up env add SESSION_SECRET=$SECRET`

# Amcrest

Unfortunately the Amcrest device can only upload via PASV so that rules out my
earlier project https://github.com/kaihendry/camftp2web since binding a range
of ports with Docker is an utter pain.

* https://amcrest.com/forum/search.php?author_id=34177&sr=posts
