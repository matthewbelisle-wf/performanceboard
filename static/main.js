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
    // TODO: Make data a dictionary {'namespace': data}
    var data = [[]];

    var fetchGraphData = function(url) {
        $.get(url, function(result) {
            while(data[0].length > 0) {
                data[0].pop();
            }
            for (i = 0; i < result.length; i++) {
                var start = Date.parse(result[i].start);
                var end = Date.parse(result[i].end);
                var y = end - start; // NOTE: accurate to a millisecond, no more!
                data[0].push({x: i, y: y});
            }
            graph.update();
        });
    };

    var fetchTopLevelUrl = function() {
        var url = '/api/' + getBoardKey();
        $.get(url, function(result) {
            if (result.series.length) {
                // TODO: Loop through series
                fetchGraphData(result.series[0]);
            }
        });
    };

    var graph = new Rickshaw.Graph({
        element: $('#graph').get(0),
        width: 600,
        height: 400,
        series: [
            {
                color: 'steelblue',
                data: data[0]
            }
        ]
    });

    fetchTopLevelUrl();
    graph.render();
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
