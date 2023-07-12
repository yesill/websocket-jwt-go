var server_url = "http://localhost:8000";
var messages_url = server_url + "/messages";
var socket = new WebSocket("ws://localhost:8000/ws");
var firstResponse = true;
var JWTToken = "";

socket.onopen = function(event) {
    console.log("WebSocket Connected Succesfully");
};

socket.onmessage = function(event) {
    var message = event.data;
    console.log("server response:", message);
    if (firstResponse) {
        JWTToken = message
        firstResponse = false;
    }
};

socket.onclose = function(event) {
    console.log("WebSocket connection ended");
};

function sendMessage() {
    const message = $('#messageInput').val();
    /* console.log("message:", message) */
    var blah = JSON.stringify(parseMessage(message))
    console.log(blah)
    $.ajax({
        url: messages_url,
        type: 'POST',
        headers: {'Token':JWTToken},
        /* dataType: 'text',
        data: {message:message}, */
        dataType: 'html',
        contentType: 'application/json',
        data: blah,
        success: function(response) {
            //console.log(response)
        },
        error: function(xhr, status, error) {
            console.error('Error sending message:', error);
        }
    });
}

function parseMessage(raw_message) {
    /* console.log("raw_message: ", raw_message) */
    raw_message = raw_message.toString();
    if (raw_message.slice(0,3) === "/to"){
        player_id = raw_message.split(" ")[1],
        message = raw_message.slice(player_id.length + 5)    // +5 stands for /to + 2 spaces
        return {
            "Type": 1,
            "PlayerID": player_id,
            "Message" : message
        }
    }
    else {
        return {
            "Type": 0,
            "PlayerID" : null,
            "Message" : raw_message
        }
    }
}