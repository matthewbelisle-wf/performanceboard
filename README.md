PerformanceBoard
================

Collects performance metrics and displays them for real-time analysis

Simple Usage
------------

Create a board if you don't have your own.

* Go to http://performanceboard-public.appspot.com/
* Click `+ New Board`

The API for the board will appear on the right side of the navbar.  Post some metrics to the API.
See the [Metrics](#metrics) section for more info on a metric's structure.

```json
{
  "key": "hotspot",
  "start": 1401763961.864,
  "end": 1401763962.567
}
```

Watch the metric show up on the board in real-time.

Metrics<a name="metrics"></a>
-------

A metric is an object with `key`, `start`, `end`, and optionally `children`.

```json
{
  "key": "hotspot",
  "start": 1401763961.864,
  "end": 1401763962.567
}
```

By specifying `children`, metrics can be nested arbitrarily deep.

```json
{
  "key": "hotspot",
  "start": 1401763961.864,
  "end": 1401763972.567,
  "children": [
    {
      "key": "sub_hotspot1",
      "start": 1401763961.864,
      "end": 1401763962.567
    },
    {
      "key": "sub_hotspot2",
      "start": 1401763963.864,
      "end": 1401763964.567,
      "children": [
         {
           "key": "sub_sub_hostspot",
           "start": 1401763963.964,
           "end": 1401763964.467
         }
      ]
    }
  ]
}
```

Plugins
-------

Any language with a REST client can use PerformanceBoard by making posts to the API, but in order
to make things easier there are some plugins for popular languages.

* [performanceboard-py](https://github.com/mgbelisle/performanceboard-py)
* [performanceboard-js](https://github.com/mgbelisle/performanceboard-js)

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
