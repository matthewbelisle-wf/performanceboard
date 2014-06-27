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

    var initGraph = function(namespace) {
        var graphElement =$('<div class="graph">');
        $('#graphs-block').append(graphElement);
        graphs[namespace.name] = graphElement.epoch({
            type: 'bar',
            data: [{
                label: namespace.name,
                values: []
            }]
        });
        updateGraph(namespace);
        setInterval(function() { updateGraph(namespace); }, 2000);
    };

    var updateGraph = function(namespace) {
        $.get(namespace.api).
            done(function(result) {
                var values = [];
                for (i = 0; i < result.length; i++) {
                    var start = Date.parse(result[i].start);
                    var end = Date.parse(result[i].end);
                    var y = end - start; // NOTE: accurate to a millisecond, no more!
                    values.push({x: i, y: y});
                }
                graphs[namespace.name].update([{
                    label: namespace.name,
                    values: values
                }]);
            });
    };

    $.get('/api/' + getBoardKey())
        .done(function(result) {
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
