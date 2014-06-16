PerformanceBoard
================

Collects performance metrics and displays them for real-time analysis

Simple Usage
------------

Create a board if you don't have your own.

* Go to http://performanceboard-public.appspot.com/
* Click `+ New Board`

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

Watch the metrics show up on the board.

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
```

Ignore the errors about [unrecognized import paths for appengine](http://stackoverflow.com/questions/22674307/go-get-package-appengine-unrecognized-import-path-appengine).

```
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
