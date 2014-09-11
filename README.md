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
  "namespace": "hotspot",
  "start": "2014-06-19T01:10:20.345287Z",
  "end": "2014-06-19T01:10:21.345287Z"
}
```

Watch the metric show up on the board in real-time.

Metrics<a name="metrics"></a>
-------

A metric is an object with `namespace`, `start`, `end`, and optionally `meta` and `children`.
Timestamps like `start` and `end` must be strings formatted according to [RFC 3339](http://www.ietf.org/rfc/rfc3339.txt).

```json
{
  "namespace": "hotspot",
  "start": "2014-06-19T01:10:20.345287Z",
  "end": "2014-06-19T01:10:21.345287Z"
}
```

By specifying `children`, metrics can be nested arbitrarily deep.

```json
{
  "namespace": "hotspot",
  "start": "2014-06-19T01:10:20.345287Z",
  "end": "2014-06-19T01:10:30.345287Z",
  "meta": {
      "mem_footprint": "10MB",
      "hostname": "foobar.com"
  },
  "children": [
    {
      "namespace": "sub_hotspot1",
      "start": "2014-06-19T01:10:21.345287Z",
      "end": "2014-06-19T01:10:22.345287Z"
    },
    {
      "namespace": "sub_hotspot2",
      "start": "2014-06-19T01:10:23.345287Z",
      "end": "2014-06-19T01:10:24.345287Z",
      "children": [
         {
           "namespace": "sub_sub_hostspot",
           "start": "2014-06-19T01:10:23.445287Z",
           "end": "2014-06-19T01:10:23.545287Z"
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

Dev Server
----------

Serve the app locally from the root directory.

```
$ goapp get github.com/mgbelisle/performanceboard/app
$ cd $GOPATH/src/github.com/mgbelisle/performanceboard
```

Ignore the errors about [unrecognized import paths for appengine](http://stackoverflow.com/questions/22674307/go-get-package-appengine-unrecognized-import-path-appengine).

```
$ goapp serve ./app
```

https://developers.google.com/appengine/docs/go/tools/devserver

Deploy
------

Deploy the app to GAE from the root directory.

```
$ goapp deploy ./app
```

https://developers.google.com/appengine/docs/go/tools/uploadinganapp
