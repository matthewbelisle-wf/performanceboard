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

///////////
// Views //
///////////

if (getBoardKey()) {
    $('#create-board-block').hide();
} else {
    $('#create-board-block').show();
}

if (getBoardKey()) {
    var api = window.location.origin + '/api/' + getBoardKey();
    $('#api').attr('href', api);
    $('#api').show();
} else {
    $('#api').hide();    
}

if (getBoardKey()) {
    $('#chart-block').show();
    var graph = new Rickshaw.Graph({
        element: $('#chart').get(0),
        width: 600,
        height: 400,
        series: [
            {
                color: 'steelblue',
                data: [ 
                    {x: 0, y: 40}, 
                    {x: 1, y: 49}, 
                    {x: 2, y: 38}, 
                    {x: 3, y: 30}, 
                    {x: 4, y: 32}
                ]
            }
        ]
    });
    graph.render();
} else {
    $('#chart-block').hide();
}
