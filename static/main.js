/////////////////
// URL helpers //
/////////////////

var getBoardKey = function() {
    return window.location.pathname.split('/')[1] || null;
};

var getMetricKeys = function() {
    return window.location.pathname.split('/').slice(2);
};

String.prototype.replaceAll = function(find, replace) {
    return this.replace(new RegExp(find, 'g'), replace);
}

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
        var xLabels = [];
        $.each(metrics, function(i, metric) {
            xLabels.push(metric.start);
            var start = Date.parse(metric.start) / 1000;
            var end = Date.parse(metric.end) / 1000;
            var x = metrics.length - i - 1;
            // var x = start;
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
        
        $('#graphs-block')
            .append($('<h2 class="graph">').text(namespace));

        var graphWrap = $('<div class="graph-wrap">');
        $('#graphs-block').append(graphWrap);

        var graphElement = $('<div class="graph">');
        graphWrap.append(graphElement);
        var graph = new Rickshaw.Graph({
            element: graphElement.get(0),
            width: 600,
            height: 400,
            renderer: 'bar',
            series: series
        });

        var xAxisElement = $('<div class="graph x-axis">');
        graphWrap.append(xAxisElement);

        var xAxis = new Rickshaw.Graph.Axis.X({
            element: xAxisElement.get(0),
            orientation: 'bottom',
            pixelsPerTick: 200,
            graph: graph,
            ticksTreatment: 'glow',
            tickFormat: function(pos) {return xLabels[pos];},
            tickRotation: 90,
            tickOffsetX: -10,
        });

        var yAxis = new Rickshaw.Graph.Axis.Y({
            graph: graph,
            ticksTreatment: 'glow'
        });

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
        // ul.append();
        var linkTemplate = 
            '<table style="text-indent:25px; width:80%">' +
            '<tr>' +
            '  <td><a href="' + v.url + '">' + v.name + '</a></td>' +
            '  <td><a href="{url}/day">Day </a></td>' +
            '  <td><a href="{url}/hour">Hour </a></td>' +
            '  <td><a href="{url}/minute">Min </a></td>' +
            '  <td><a href="{url}/second">Sec </a></td>' +
            '</tr>' +
            '</table>';
        ul.append(linkTemplate.replaceAll("{url}", v.url));
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
