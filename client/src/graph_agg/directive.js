var $ = require('jquery');
var Rickshaw = require('rickshaw');
var angular = require('angular');
var fs = require('fs');
var d3 = require('d3');

var directive = function(
    $http
) {
    var initGraph = function(index, data) {
        var binStats = data.results;
        var xLabels = [];
        var seriesMap = {
            max: {
                data: [],
                name: 'max',
                color: 'red'
            },
            mean: {
                data: [],
                name: 'mean',
                color: 'green'
            },
            min: {
                data: [],
                name: 'min',
                color: 'blue'
            },
        };

        $.each(binStats, function(i, binStat) {
            xLabels.push(binStat.start);
            var start = Date.parse(binStat.start) / 1000;
            var x = binStats.length - i - 1;

            seriesMap.min.data.unshift({x: x, y: binStat.min});
            seriesMap.mean.data.unshift({x: x, y: binStat.mean});
            seriesMap.max.data.unshift({x: x, y: binStat.max});
        });

        var series = 
        [
            seriesMap.min, 
            seriesMap.mean,
            seriesMap.max
        ];

        var graph = new Rickshaw.Graph({
            element: $('#graph-' + index).get(0),
            height: 400,
            renderer: 'line',
            series: series
        });

        var xAxis = new Rickshaw.Graph.Axis.X({
            element: $('#x-axis-' + index).get(0),
            orientation: 'bottom',
            pixelsPerTick: 200,
            graph: graph,
            ticksTreatment: 'glow',
            tickFormat: function(pos) {return xLabels[pos];},
            tickRotation: 90,
            tickOffsetX: -10,
        });

        new Rickshaw.Graph.Axis.Y({
            graph: graph,
            ticksTreatment: 'glow',
            tickFormat: function(y) { return y + 'ms'; }
        });

        new Rickshaw.Graph.Axis.Time({
          graph: graph
        });

        new Rickshaw.Graph.HoverDetail({
          graph: graph
        });

        graph.render();
    };

    return {
        restrict: 'E',
        template: fs.readFileSync(__dirname + '/template.html', 'utf8'),
        link: function(scope, element, attrs) {
            scope.name = attrs.name;
            scope.index = attrs.index;
            var path = window.location.pathname;
            var aggregate = '/second';
            if(path.endsWith('minute'))
                aggregate = '/minute';
            if(path.endsWith('hour'))
                aggregate = '/hour';
            if(path.endsWith('day'))
                aggregate = '/day';

            var api = attrs.api + aggregate;
            $http({method: 'GET', url: api}).
                success(function(data) {
                    initGraph(attrs.index, data);
                });
        }
    };
};
directive.$inject = [
    '$http'
];

module.exports = directive;
