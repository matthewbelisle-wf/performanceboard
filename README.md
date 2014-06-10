PerformanceBoard
================

Collects performance metrics and displays them for real-time analysis

Simple Usage
------------

Create a board if you don't have your own.

```
$ curl -X POST -d '' http://performanceboard-public.appspot.com/api/
{
  "api": "http://performanceboard-public.appspot.com/api/ahlzfnBlcmZvcm1hbmNlYm9hcmQtcHVibGljchILEgVCb2FyZBiAgICAmc6UCgw",
  "client": "http://performanceboard-public.appspot.com/ahlzfnBlcmZvcm1hbmNlYm9hcmQtcHVibGljchILEgVCb2FyZBiAgICAmc6UCgw",
  "key": "ahlzfnBlcmZvcm1hbmNlYm9hcmQtcHVibGljchILEgVCb2FyZBiAgICAmc6UCgw"
}
```

Post some metrics to the board.

```
$ cat metrics.json
[
  {
    "key": "hotspot",
    "start": 1401763961.864,
    "end": 1401763962.567
  }
]
$ curl -X POST -d @metrics.json http://performanceboard-public.appspot.com/api/ahlzfnBlcmZvcm1hbmNlYm9hcmQtcHVibGljchILEgVCb2FyZBiAgICAmc6UCgw
```

Open the board in a browser.  (TODO: Actually make the client)

http://performanceboard-public.appspot.com/ahlzfnBlcmZvcm1hbmNlYm9hcmQtcHVibGljchILEgVCb2FyZBiAgICAmc6UCgw

Plugins
-------

Any language with a REST client can use PerformanceBoard by making posts to the API, but in order
to make things easier there are some plugins for popular languages.

* [performanceboard-py](https://github.com/mgbelisle/performanceboard-py)

Dev Server
----------

Serve the app locally from the root directory.

```
$ go get ./app/
$ goapp serve .
```

https://developers.google.com/appengine/docs/go/tools/devserver

Deploy
------

Deploy the app to GAE from the root directory.

```
$ goapp deploy .
```

https://developers.google.com/appengine/docs/go/tools/uploadinganapp
