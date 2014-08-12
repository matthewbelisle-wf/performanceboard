// <script src="//ajax.googleapis.com/ajax/libs/jquery/1.11.1/jquery.min.js"></script>


var PerformanceboardService = function(){

    var endPoint = '{{.post_url}}';

    var metrics = {};
    var nextKey = 0;

    /**
     * Retrieves a single PDF file by keyname.
     * @param name identifies this measurment for visualization
     *
     * @param key specify an active metric to attach a measurement.
     *            pass null/undefined/false to create a root metric
     *            a metric is stateful, so calling start repeatedly
     *            with the same key generates a hierarchy of children
     * @param metadata (optional) an object with metadata for this timing measurement
     * @return the key param, or a new key if none was provided
     */
    var start = function(name, key, metadata) {
        var newMetric = {
            namespace: name,
            start: new Date().toISOString()
        };

        if(metadata) {
            newMetric.meta = metadata;
        }

        if(!key) {
            key = nextKey;
            nextKey = nextKey + 1;
        }

        if(metrics.key) {
            metrics.key.stack.push(newMetric);
        }
        else {
            metrics.key = {
                stack: [newMetric]
            };
        }
        return key;
    };

    /**
     * Sets the stop time on a metric. Posts to the server if it's for a root measurment.
     * @param key specify a metric to finish.
     * @param metadata (optional) additional data to extend the metadata set on start()
     * @return the depth of the metrics tree after this call. if zero, the
     *         metric has been posted and the key has become invalid
     */
    var stop = function(key, metadata) {
        var stack = metrics.key.stack;
        if(!stack) {
            return -1;
        }

        var metric = stack.pop();
        metric.stop = new Date().toISOString();
        if(metadata) {
            $.extend(metric.meta, metadata);
        }

        var depth = stack.length;
        if(depth === 0) {
            $.ajax({
                crossDomain: true,
                type: "POST",
                url: endPoint,
                data: JSON.stringify(metric),
                dataType: 'json',
                complete: function(){}
            });
            delete metrics.key;
        } else {
            var parent = stack.pop();
            if(!parent.children) {
                parent.children = [];
            }
            parent.children.push(metric);
            stack.push(parent);
        }
        return depth;
    };

    return {
        start: start,
        stop: stop
    };
};
window._pb = PerformanceboardService;
