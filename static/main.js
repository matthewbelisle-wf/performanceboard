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
    var initGraph = function(namespace) {
        var titleElement = $('<h1 class="graph-title">').text(namespace.name);
        $('#graphs-block').append(titleElement);
        var graphElement = $('<div class="graph">');
        $('#graphs-block').append(graphElement);
        new Rickshaw.Graph.Ajax({
            dataURL: namespace.api,
            element: graphElement.get(0),
            width: 600,
            height: 400,
            renderer: 'bar',
            onData: onData
        });
    };

    var onData = function(data) {
        var series = [{data: [], color: 'lightblue'}];
        var metrics = data.result;
        for (i = 0; i < metrics.length; i++) {
            var start = Date.parse(metrics[i].start) / 1000;
            var end = Date.parse(metrics[i].end) / 1000;
            var y = end - start; // NOTE: accurate to a millisecond, no more!
            series[0].data.unshift({x: metrics.length - i - 1, y: y});
        }
        console.log(JSON.stringify(series));
        return series;
    };

    $.get('/api/' + getBoardKey())
        .done(function(data) {
            for (var i = 0; i < data.namespaces.length; i++) {
                initGraph(data.namespaces[i]);
            }
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
