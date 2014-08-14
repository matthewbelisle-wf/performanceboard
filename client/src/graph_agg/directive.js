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
                renderer: 'line',
                name: 'max',
                color: 'red'
            },
            mean: {
                data: [],
                renderer: 'line',
                name: 'mean',
                color: 'green'
            },
            min: {
                data: [],
                renderer: 'line',
                name: 'min',
                color: 'blue'
            },
            count: {
                data: [],
                renderer: 'stack',
                name: 'counts',
                color: 'grey'
            },
        };

        var min = -1;
        var max = -1;
        var maxCount = 0;

        $.each(binStats, function(i, binStat) {
            xLabels.push(binStat.start);
            var start = Date.parse(binStat.start) / 1000;
            var x = binStats.length - i - 1;
            if(min === -1 || min > binStat.min)
                min = binStat.min;
            if(max === -1 || max < binStat.max)
                max = binStat.max;
            if(maxCount < binStat.count)
                maxCount = binStat.count;

            seriesMap.min.data.unshift({x: x, y: binStat.min});
            seriesMap.mean.data.unshift({x: x, y: binStat.mean});
            seriesMap.max.data.unshift({x: x, y: binStat.max});
            seriesMap.count.data.unshift({x: x, y: binStat.count});
        });

        var scale = d3.scale.linear().domain([0, max]).nice();
        seriesMap.max.scale = scale;
        seriesMap.mean.scale = scale;
        seriesMap.min.scale = scale;

        var scale2 = d3.scale.linear().domain([0, maxCount]).nice();
        seriesMap.count.scale = scale2;

        var series = 
        [
            seriesMap.count,
            seriesMap.max,
            seriesMap.mean,
            seriesMap.min 
        ];

        var graph = new Rickshaw.Graph({
            element: $('#graph-' + index).get(0),
            height: 400,
            renderer: 'multi',
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

        // TODO get rid of this
        // new Rickshaw.Graph.Axis.Y({
        //     graph: graph,
        //     ticksTreatment: 'glow',
        //     tickFormat: function(y) { return y + 'ms'; }
        // });

        // TODO hang this on an HTML element
        new Rickshaw.Graph.Axis.Y.Scaled({
          graph: graph,
          orientation: 'right',
          scale: scale,
          tickFormat: function(y) { return y + 'ms'; }
        });

        // TODO hang this on an HTML element
        // new Rickshaw.Graph.Axis.Y.Scaled({
        //   graph: graph,
        //   orientation: 'right',
        //   scale: scale2,
        // });

        new Rickshaw.Graph.Axis.Time({
          graph: graph
        });

        new Rickshaw.Graph.Legend({
            graph: graph,
            element: $('#legend-' + index).get(0),
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
