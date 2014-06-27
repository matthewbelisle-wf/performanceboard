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
    var graphs = {}; // {'namespace.name': graph}
    var data = {}; // {'namespace.name': data}

    var initGraph = function(namespace) {
        var graphElement = $('<div>');
        $('#graphs-block').append(graphElement);
        data[namespace.name] = [];
        updateData(namespace);
        graphs[namespace.name] = new Rickshaw.Graph({
            element: graphElement.get(0),
            width: 600,
            height: 400,
            series: [{data: data[namespace.name], color: 'black'}]
        });
        graphs[namespace.name].render();
        console.log(graphs[namespace.name]);
    };

    var updateData = function(namespace) {
        $.get(namespace.api, function(result) {
            data[namespace.name].length = 0;
            for (i = 0; i < result.length; i++) {
                var start = Date.parse(result[i].start);
                var end = Date.parse(result[i].end);
                var y = end - start; // NOTE: accurate to a millisecond, no more!
                data[namespace.name].push({x: i, y: y});
            }
            graphs[namespace.name].update();
        });
    };

    $.get('/api/' + getBoardKey())
        .success(function(result) {
            for (var i = 0; i < result.namespaces.length; i++) {
                initGraph(result.namespaces[i]);
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
