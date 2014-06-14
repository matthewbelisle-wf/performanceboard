$('#create-board').click(function() {
    $.post('/api/')
        .done(function(data) {
            document.location.href = data.client;
        });
});
