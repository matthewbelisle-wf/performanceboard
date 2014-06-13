$('#create-board').click(function() {
    $.post('/api/')
        .done(function(data) {
            console.log(data);
        });
});
