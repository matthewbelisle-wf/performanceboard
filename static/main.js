/////////////////
// URL helpers //
/////////////////

var getBoardKey = function() {
    return window.location.pathname.split('/')[1] || null;
};

var getMetricKeys = function() {
    return window.location.pathname.split('/').slice(2);
};

/////////////
// Buttons //
/////////////

$('#create-board').click(function() {
    $.post('/api/')
        .done(function(data) {
            // TODO: history.pushState() maybe?
            window.location.href = '/' + data.board;
        });
});

////////////
// Graphs //
////////////

var initGraphs = function() {
    var initGraph = function(namespace, data) {
        var palette = new Rickshaw.Color.Palette();
        var seriesMap = {};
        var metrics = data.result;
        for (var i = 0; i < metrics.length; i++) {
            var metric = metrics[i];
            var start = Date.parse(metric.start) / 1000;
            var end = Date.parse(metric.end) / 1000;
            var x = metrics.length - i - 1;
            var y = end - start; // NOTE: accurate to a millisecond, no more!
            for (var i2 = 0; metric.children && i2 < metric.children.length; i2++) {
                var child = metric.children[i2];
                var start2 = Date.parse(child.start) / 1000;
                var end2 = Date.parse(child.end) / 1000;
                var y2 = end2 - start2;
                y -= y2;
                if (!seriesMap[child.namespace]) {
                    seriesMap[child.namespace] = {
                        data: [],
                        name: child.namespace,
                        color: palette.color()
                    };
                }
                seriesMap[child.namespace].data.unshift({x: x, y: y2});
            }
            if (!seriesMap[metric.namespace]) {
                seriesMap[metric.namespace] = {
                    data: [],
                    name: metric.namespace,
                    color: palette.color()
                };
            }
            seriesMap[metric.namespace].data.unshift({x: x, y: y});
        }

        var series = [];
        for (var n in seriesMap) {
            series.push(seriesMap[n]);
        }
        // graph.series = series;

        var yAxis = new Rickshaw.Graph.Axis.Y({
            graph: graph,
            ticksTreatment: 'glow'
        });

        yAxis.render();

        var titleElement = $('<h1 class="graph-title">').text(namespace.name);
        $('#graphs-block').append(titleElement);

        var graphElement = $('<div class="graph">');
        $('#graphs-block').append(graphElement);
        var graph = new Rickshaw.Graph.Ajax({
            dataURL: namespace.api + '?depth=1',
            element: graphElement.get(0),
            width: 600,
            height: 400,
            renderer: 'bar',
            onData: onData
        });

        // var previewElement = $('<div class="preview">');
        // graphElement.append(previewElement);
        // var preview = new Rickshaw.Graph.RangeSlider({
        //     graph: graph,
        //     element: previewElement.get(0)
        // });        
    };

    $.get('/api/' + getBoardKey())
        .done(function(data) {
            $.each(data.namespaces, function(i, namespace) {
                $.get(namespace.api, {depth: 1})
                    .done(function(data2) {
                        initGraph(namespace.name, data2);
                    });
            });
        });
};

///////////
// Views //
///////////

if (getBoardKey()) {
    $('#create-board-block').hide();
} else {
    $('#create-board-block').show();
    $.get('/api/').done(function(data) {
        //TODO:: replace with a template
        var html = '<ul>';
        for (var i = 0; i < data.results.length; i++) {
            html += '<li><a href=' + data.results[i].url + '>' + data.results[i].name + '</a></li>'
        }
        html += '</ul>'
        $('#list-boards-block').html(html);
    })
}

if (getBoardKey()) {
    var api = '/api/' + getBoardKey();
    $('#api').attr('href', api);
    $('#api').show();
} else {
    $('#api').hide();  
}

if (getBoardKey()) {
    initGraphs();
    $('#graph-block').show();
} else {
    $('#graph-block').hide();
}
