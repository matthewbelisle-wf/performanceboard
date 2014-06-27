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
            type: 'time.bar',
            data: [{
                label: namespace.name,
                values: [{
                    time: Date.now() / 1000,
                    y: 0
                }]
            }]
        });
        updateGraph(namespace);
        setInterval(function() { updateGraph(namespace); }, 2000);
    };

    var updateGraph = function(namespace) {
        $.get(namespace.api).
            done(function(data) {
                for (i = 0; i < data.length; i++) {
                    var start = Date.parse(data[i].start) / 1000;
                    var end = Date.parse(data[i].end) / 1000;
                    var y = end - start; // NOTE: accurate to a millisecond, no more!
                    graphs[namespace.name].push([{time: start, y: y}]);
                    console.log(data[i]);
                    console.log({time: start, y: y});
                }
            });
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
