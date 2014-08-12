var $ = require('jquery');
var Rickshaw = require('rickshaw');
var angular = require('angular');
var fs = require('fs');

var directive = function(
    $http
) {
    var initGraph = function(index, data) {
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
        
        console.log(angular.element('#graph-' + index)[0]);
        console.log(series);
        var graph = new Rickshaw.Graph({
            element: angular.element('#graph-' + index)[0],
            height: 400,
            renderer: 'bar',
            series: series
        });

        // var xAxisElement = $('<div class="graph x-axis">');
        // graphWrap.append(xAxisElement);

        // var xAxis = new Rickshaw.Graph.Axis.X({
        //     element: xAxisElement.get(0),
        //     orientation: 'bottom',
        //     pixelsPerTick: 200,
        //     graph: graph,
        //     ticksTreatment: 'glow',
        //     tickFormat: function(pos) {return xLabels[pos];},
        //     tickRotation: 90,
        //     tickOffsetX: -10,
        // });

        // var yAxis = new Rickshaw.Graph.Axis.Y({
        //     graph: graph,
        //     ticksTreatment: 'glow'
        // });

        // graph.render();
    };

    return {
        restrict: 'E',
        template: fs.readFileSync(__dirname + '/template.html', 'utf8'),
        link: function(scope, element, attrs) {
            scope.name = attrs.name;
            scope.index = attrs.index;
            $http({method: 'GET', url: attrs.api}).
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
