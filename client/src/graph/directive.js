var $ = require('jquery');
var jqueryUI = require('jquery-ui');
var Rickshaw = require('rickshaw');
var angular = require('angular');
var fs = require('fs');

var directive = function(
    $http,
    slider
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

        var graph = new Rickshaw.Graph({
            element: $('#graph-' + index).get(0),
            height: 400,
            renderer: 'bar',
            interpolation: 'linear',
            series: series,
        });

        var xAxis = new Rickshaw.Graph.Axis.X({
            element: $('#x-axis-' + index).get(0),
            orientation: 'bottom',
            pixelsPerTick: 200,
            graph: graph,
            ticksTreatment: 'glow',
            tickFormat: function(pos) {
                // return (new Date(pos * 1000)).toGMTString();
                return (new Date(xLabels[pos])).toGMTString();
            },
            tickRotation: 90,
            tickOffsetX: -10,
        });

        var yAxis = new Rickshaw.Graph.Axis.Y({
            graph: graph,
            ticksTreatment: 'glow'
        });

        graph.render();

        var domain = graph.dataDomain();
        var slider = $('#x-axis-slider').slider({
            range: true,
            min: domain[0],
            max: domain[1],
            values: [
                domain[0],
                domain[1]
            ],
            slide: function(event, ui) {
                if (ui.values[1] <= ui.values[0]) return;

                graph.window.xMin = ui.values[0];
                graph.window.xMax = ui.values[1];
                graph.update();

                var domain = graph.dataDomain();

                // if we're at an extreme, stick there
                if (domain[0] == ui.values[0]) {
                    graph.window.xMin = undefined;
                }
                if (domain[1] == ui.values[1]) {
                    graph.window.xMax = undefined;
                }
            }
        });

        slider.css('width', graph.width + 'px')
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
    '$http',
];

module.exports = directive;
