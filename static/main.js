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
        var metrics = data.results;
        $.each(metrics, function(i, metric) {
            var start = Date.parse(metric.start) / 1000;
            var end = Date.parse(metric.end) / 1000;
            var x = metrics.length - i - 1;
            var y = end - start; // NOTE: accurate to a millisecond, no more!
            if (metric.children) {
                $.each(metric.children, function(i2, child) {
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
                });
            }
            if (!seriesMap[metric.namespace]) {
                seriesMap[metric.namespace] = {
                    data: [],
                    name: metric.namespace,
                    color: palette.color()
                };
            }
            seriesMap[metric.namespace].data.unshift({x: x, y: y});
        });

        var series = [];
        $.each(seriesMap, function(k, v) {
            series.push(v);
        });

        var graphWrap = $('<div class="graph-wrap">');
        $('#graphs-block').append(graphWrap);        

        var titleElement = $('<h1 class="graph-title">').text(namespace);
        graphWrap.append(titleElement);

        var graphElement = $('<div class="graph">');
        graphWrap.append(graphElement);
        var graph = new Rickshaw.Graph({
            element: graphElement.get(0),
            width: 600,
            height: 400,
            renderer: 'bar',
            series: series
        });

        // var yAxisElement = $('<div class="axis y-axis">');
        // graphWrap.append(yAxisElement);
        // var yAxis = new Rickshaw.Graph.Axis.Y({
        //     element: yAxisElement.get(0),
        //     graph: graph,
        //     tickFormat: Rickshaw.Fixtures.Number.formatKMBT,
        //     ticksTreatment: 'glow'
        // });

        graph.render();

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

////////////
// Boards //
////////////

$.get('/api/').done(function(data) {
    // TODO:: replace with a template
    var ul = $('<ul class="nav nav-stacked">');
    $.each(data.results, function(k, v) {
        ul.append('<li><a href="' + v.url + '">' + v.name + '</a></li>');
    });
    $('#list-boards-block').html(ul);
});

///////////
// Views //
///////////

if (getBoardKey()) {
    $('#create-board-block').hide();
} else {
    $('#create-board-block').show();
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
